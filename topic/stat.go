package topic

import (
	"encoding/xml"
	"slices"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/matter"
)

type PageStat struct {
	kosmos.BaseModel `bson:",inline" kosmos:"page_stat"`
	XMLName          xml.Name `xml:"page_stat" json:"-" bson:"-"`
	Title            string   `xml:"title" json:"Title" bson:"title"`
	CommentCount     int      `xml:"commentcount" json:"CommentCount" bson:"comment_count,omitempty"`
	Authors          []string `xml:"authors" json:"Authors" bson:"authors,omitempty"`
	commentIncr      int
	authorsAdd       []string
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
