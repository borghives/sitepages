package topic

import (
	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RelationTopicResponse struct {
	BaseTopicResponse
	LinkDescs []sitepages.LinkDescription `xml:"-" json:"LinkDescs,omitempty" bson:"-" `
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

	return t.BaseTopicResponse.Append(data)
}

func (t *RelationTopicResponse) New() TopicResponse {
	return &RelationTopicResponse{}
}
