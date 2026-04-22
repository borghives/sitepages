package topic

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type List struct {
	ID       string          `xml:"-" json:"ID" bson:"name"`
	Contents []bson.ObjectID `xml:"-" json:"Contents" bson:"contents"`
}

type ListTopicResponse struct {
	EntangledResponse
	ListData []List `xml:"-" json:"ListData,omitempty" bson:"listdata,omitempty" `
}

func NewListTopicResponse(name string) Response {
	listName := name
	var list *List
	if listName != "" {
		list = &List{ID: listName}
	}

	var listData []List
	if list != nil {
		listData = []List{*list}
	}

	return &ListTopicResponse{
		ListData: listData,
	}
}

func (t *ListTopicResponse) Append(data any) bson.ObjectID {

	id := t.EntangledResponse.Append(data)
	if !id.IsZero() {
		if len(t.ListData) > 0 {
			t.ListData[0].Contents = append(t.ListData[0].Contents, id)
		}
	}
	return id
}
