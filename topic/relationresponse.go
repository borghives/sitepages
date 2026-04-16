package topic

import (
	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RelationTopicResponse struct {
	BaseResponse
	LinkDescs []sitepages.LinkDescription `xml:"-" json:"LinkDescs,omitempty" bson:"-" `
}

func NewRelationTopicResponse() Response {
	return &RelationTopicResponse{}
}

func (t *RelationTopicResponse) Append(data any) bson.ObjectID {
	switch response := data.(type) {
	case sitepages.UserToPageLink:
		t.LinkDescs = append(t.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case *sitepages.UserToPageLink:
		t.LinkDescs = append(t.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case sitepages.UserToCommentLink:
		t.LinkDescs = append(t.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case *sitepages.UserToCommentLink:
		t.LinkDescs = append(t.LinkDescs, response.LinkDescription)
		return response.ObjectId
	}

	return t.BaseResponse.Append(data)
}
