package topic

import (
	"encoding/xml"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/matter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PageStat struct {
	kosmos.BaseModel `bson:",inline" kosmos:"page_stat"`
	XMLName          xml.Name      `xml:"page_stat" json:"-" bson:"-"`
	Root             bson.ObjectID `xml:"root" json:"Root" bson:"root"`
	Title            string        `xml:"title" json:"Title" bson:"title"`
	CommentCount     int           `xml:"commentcount" json:"CommentCount" bson:"comment_count,omitempty"`
	Authors          []string      `xml:"authors" json:"Authors" bson:"authors,omitempty"`
	commentIncr      int
}

func (p *PageStat) IncrCommentCount() {
	p.commentIncr += 1
	p.CommentCount += 1
}

func (p *PageStat) Collapse() matter.Ripple {
	ripple := p.BaseModel.Collapse()
	ripple.Set("comment_count", p.CommentCount)
	ripple.DoIncr("comment_count", p.commentIncr)
	p.CommentCount = 0
	p.commentIncr = 0
	return ripple
}

func (p *PageStat) Decohere(ripple matter.Ripple) error {
	count, ok := ripple.Get("comment_count")
	if ok {
		p.CommentCount = count.(int)
	}

	return p.Decohere(ripple)
}
