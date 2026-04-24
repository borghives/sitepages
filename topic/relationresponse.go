package topic

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RelationTopicResponse struct {
	EntangledResponse
	LinkDescs []LinkDescription `xml:"-" json:"LinkDescs,omitempty" bson:"-" `
}

func NewRelationTopicResponse() Response {
	return &RelationTopicResponse{}
}

func (rr *RelationTopicResponse) Append(data any) bson.ObjectID {
	switch response := data.(type) {
	case UserToPageLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case *UserToPageLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case UserToCommentLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	case *UserToCommentLink:
		rr.LinkDescs = append(rr.LinkDescs, response.LinkDescription)
		return response.ObjectId
	}

	return rr.BaseResponse.Append(data)
}
