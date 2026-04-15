package topic

import (
	"fmt"
	"log"

	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TopicResponse interface {
	New() TopicResponse
	SetOnError(err error, code int) error
	HasError() bool
	GetStatus() StatusResponse
	SetTargetId(id bson.ObjectID)
	GetTargetId() *bson.ObjectID
	Append(data any) bson.ObjectID
}

type StatusResponse struct {
	StatusCode int    `xml:"-" json:"-" bson:"-" `
	StatusMsg  string `xml:"-" json:"error,omitempty" bson:"-" `
}

func (e StatusResponse) GetStatus() StatusResponse {
	return e
}

func (e StatusResponse) Error() string {
	return fmt.Sprintf("Response Status %d: %s", e.StatusCode, e.StatusMsg)
}

func (e *StatusResponse) SetOnError(err error, code int) error {
	if err != nil {
		log.Printf("Response Error: %s", err.Error())
		if !e.HasError() {

			//Personal rule: only set the first error
			e.StatusMsg = err.Error()
			e.StatusCode = code
		}

		return e
	}
	return nil
}

func (e *StatusResponse) HasError() bool {
	return e.StatusCode >= 400
}

type BaseTopicResponse struct {
	StatusResponse
	TargetId    *bson.ObjectID       `xml:"-" json:"TargetId,omitempty" bson:"targetid,omitempty" `
	PageData    []sitepages.SitePage `xml:"-" json:"PageData,omitempty" bson:"pagedata,omitempty" `
	StanzaData  []sitepages.Stanza   `xml:"-" json:"StanzaData,omitempty" bson:"stanzadata,omitempty" `
	CommentData []sitepages.Comment  `xml:"-" json:"CommentData,omitempty" bson:"commentdata,omitempty" `
}

func (t *BaseTopicResponse) New() TopicResponse {
	return &BaseTopicResponse{}
}

func (t *BaseTopicResponse) SetTargetId(id bson.ObjectID) {
	t.TargetId = &id
}

func (t *BaseTopicResponse) GetTargetId() *bson.ObjectID {
	return t.TargetId
}

func (t *BaseTopicResponse) Append(data any) bson.ObjectID {
	var id bson.ObjectID
	switch response := data.(type) {
	case *sitepages.SitePage:
		t.PageData = append(t.PageData, *response)
		id = response.ID
	case *sitepages.Stanza:
		t.StanzaData = append(t.StanzaData, *response)
		id = response.ID
	case *sitepages.Comment:
		t.CommentData = append(t.CommentData, *response)
		id = response.ID
	case sitepages.SitePage:
		t.PageData = append(t.PageData, response)
		id = response.ID
	case sitepages.Stanza:
		t.StanzaData = append(t.StanzaData, response)
		id = response.ID
	case sitepages.Comment:
		t.CommentData = append(t.CommentData, response)
		id = response.ID
	}

	return id
}
