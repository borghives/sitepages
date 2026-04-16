package topic

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type List struct {
	ID       string          `xml:"-" json:"ID" bson:"name"`
	Contents []bson.ObjectID `xml:"-" json:"Contents" bson:"contents"`
}

type ListTopicResponse struct {
	BaseResponse
	ListData []List `xml:"-" json:"ListData,omitempty" bson:"listdata,omitempty" `
}

func (t *ListTopicResponse) Append(data any) bson.ObjectID {

	id := t.BaseResponse.Append(data)
	if !id.IsZero() {
		if len(t.ListData) > 0 {
			t.ListData[0].Contents = append(t.ListData[0].Contents, id)
		}
	}
	return id
}

func (t *ListTopicResponse) New() Response {
	return &ListTopicResponse{}
}
