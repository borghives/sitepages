package topic

import (
	"cmp"
	"context"
	"encoding/xml"
	"math"
	"slices"
	"time"

	km "github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/matter"
	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PageStat struct {
	km.BaseModel `bson:",inline" kosmos:"page_stat"`
	XMLName      xml.Name `xml:"page_stat" json:"-" bson:"-"`
	Title        string   `xml:"title" json:"title" bson:"title"`
	CommentCount int      `xml:"commentcount" json:"commentcount" bson:"comment_count,omitempty"`
	Authors      []string `xml:"authors" json:"authors" bson:"authors,omitempty"`
	commentIncr  int
	authorsAdd   []string
}

func (p *PageStat) IncrCommentCount() {
	p.commentIncr += 1
	p.CommentCount += 1
}

func (p *PageStat) AddUniqueAuthor(author string) {
	if author == "" {
		return
	}

	if slices.Contains(p.Authors, author) {
		return // Author already exists
	}
	p.authorsAdd = appendUnique(p.authorsAdd, author)

}

func (p *PageStat) Collapse() matter.Ripple {
	ripple := p.BaseModel.Collapse()
	ripple.Set("comment_count", p.CommentCount)
	ripple.Set("authors", append(p.Authors, p.authorsAdd...))

	ripple.DoIncr("comment_count", p.commentIncr)

	for _, author := range p.authorsAdd {
		ripple.DoAddToSet("authors", author)
	}

	p.CommentCount = 0
	p.Authors = nil
	return ripple
}

func (p *PageStat) Decohere(ripple matter.Ripple) error {
	count, ok := ripple.Get("comment_count")
	if ok {
		p.CommentCount = count.(int)
	}
	p.commentIncr = 0

	authors, ok := ripple.Get("authors")
	if ok {
		p.Authors = authors.([]string)
	}

	p.authorsAdd = nil

	return p.BaseModel.Decohere(ripple)
}

func appendUnique(slice []string, element string) []string {
	if slices.Contains(slice, element) {
		return slice // Element already exists, return the original slice
	}
	return append(slice, element) // Element is new, append and return
}

func (p PageStat) FreshnessScore() float32 {
	maxScore := float32(p.CommentCount) + 2.0*float32(len(p.Authors))

	now := time.Now()

	// Guard against negative age (clock skew or future dates)
	if p.UpdatedTime.After(now) {
		return maxScore
	}

	ageHours := now.Sub(p.UpdatedTime).Hours()

	// Calculate lambda: ln(2) / half_life
	halfLifeHours := 24.0
	lambda := math.Ln2 / halfLifeHours
	// Calculate exponential decay multiplier: e^(-lambda * age)
	decayMultiplier := math.Exp(-lambda * ageHours)
	return maxScore * float32(decayMultiplier)

}

func TopFreshPageRoot(ctx context.Context) ([]PageStat, error) {
	pageStats, err := km.Detect[PageStat]().SortLatest().Limit(200).PullAll(ctx)

	if err != nil {
		return nil, err
	}

	slices.SortFunc(pageStats, func(a, b PageStat) int {
		scoreA := a.FreshnessScore()
		scoreB := b.FreshnessScore()
		return cmp.Compare(scoreB, scoreA)
	})

	return pageStats, nil
}

type RelationCount struct {
	Relation string `bson:"relation"`
	Count    int    `bson:"count"`
}

func (r RelationCount) GetID() bson.ObjectID {
	return bson.NilObjectID
}

func (r RelationCount) HasID() bool {
	return false
}

func (r RelationCount) LastObserved() time.Time {
	return time.Time{}
}

type PageRank struct {
	sitepages.Page `bson:",inline"`
	score          float32
	hasScore       bool
}

func (p *PageRank) LinkScore(ctx context.Context) float32 {
	if p.hasScore {
		return p.score
	}

	p.hasScore = true

	relationCounts, err := km.ProjectInto[RelationCount](
		km.Fld("relation").With().GroupKey("relation"),
		km.Fld("count").With().One(),
	).From(
		km.Detect[UserToPageLink](
			km.Fld("objid").ID().Eq(p.Page.ID),
			km.Fld("state").Eq("active"),
		).GroupBy(
			"relation",
		).Accumulate(
			km.Fld("count").With().Sum(1),
		),
	).PullAll(ctx)

	if err != nil {
		return 0.0
	}

	var finalScore float32

	for _, relation := range relationCounts {
		switch relation.Relation {
		case "bookmarked":
			finalScore += 2.0
		case "endorsed":
			finalScore += 1.5
		default:
			finalScore += 1.0
		}
	}

	p.score = finalScore

	return p.score
}

func TopPageByRoot(ctx context.Context, root bson.ObjectID) (*PageRank, error) {
	pages, err := km.Detect[PageRank](
		km.Fld("Root").ID().Eq(root),
	).SortLatest().Limit(100).PullAll(ctx)

	if len(pages) == 0 {
		return nil, nil
	}

	slices.SortFunc(pages, func(a, b PageRank) int {
		scoreA := a.LinkScore(ctx)
		scoreB := b.LinkScore(ctx)
		return cmp.Compare(scoreB, scoreA)
	})

	if err != nil {
		return nil, err
	}

	return &pages[0], nil
}

// func SelectPagesByRoot(ctx context.Context, roots []bson.ObjectID) (pages []Page, err error) {
// 	pageStats, err := km.Detect[sitepages.Page](
// 		km.Fld("Root").ID().In(roots),
// 	).Sort().GroupBy("Root").Limit(200).PullAll(ctx)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return
// }
