package topic

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type List struct {
	ID        string            `xml:"-" json:"ID" bson:"name"`
	Contents  []bson.ObjectID   `xml:"-" json:"Contents" bson:"contents"`
	LinkDescs []LinkDescription `xml:"-" json:"LinkDescs,omitempty" bson:"linkdescs,omitempty" `
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

func (lr *ListTopicResponse) Append(data any) bson.ObjectID {
	var id bson.ObjectID
	var link *LinkDescription
	switch data := data.(type) {
	case UserToPageLink:
		id = data.ObjectId
		link = &data.LinkDescription
	case UserToCommentLink:
		id = data.ObjectId
		link = &data.LinkDescription
	default:
		id = lr.EntangledResponse.Append(data)
	}
	if len(lr.ListData) > 0 {
		if !id.IsZero() {
			lr.ListData[0].Contents = append(lr.ListData[0].Contents, id)
		}
		if link != nil {
			lr.ListData[0].LinkDescs = append(lr.ListData[0].LinkDescs, *link)
		}
	}
	return id
}
