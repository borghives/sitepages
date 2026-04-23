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

func (rr *RelationTopicResponse) Append(data any) bson.ObjectID {
	switch response := data.(type) {
	case sitepages.UserToPageLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case *sitepages.UserToPageLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case sitepages.UserToCommentLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case *sitepages.UserToCommentLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	}

	return rr.BaseResponse.Append(data)
}
