package sitepages

import (
	"github.com/borghives/kosmos-go"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type LinkDescription struct {
	SubjectId bson.ObjectID `json:"-" bson:"subjid"`
	ObjectId  bson.ObjectID `json:"ObjId" bson:"objid"`
	Relation  string        `json:"Relation" bson:"relation"`
	CreatorId bson.ObjectID `json:"-" bson:"creatorid"`
	Name      string        `json:"-" bson:"name,omitempty"`
	State     string        `json:"State,omitempty" bson:"state,omitempty"`
}

type UserToPageLink struct {
	kosmos.BaseModel `bson:",inline" kosmos:"user_page"`
	LinkDescription  `bson:",inline"`
}

type UserToCommentLink struct {
	kosmos.BaseModel `bson:",inline" kosmos:"rel_commentrelation"`
	LinkDescription  `bson:",inline"`
}
