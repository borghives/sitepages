package topic

import (
	"fmt"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/matter/expression"
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

func (ld *LinkDescription) SetSubjectID(id bson.ObjectID) {
	ld.SubjectId = id
}

func (ld *LinkDescription) SetObjectID(id bson.ObjectID) {
	ld.ObjectId = id
}

func (ld *LinkDescription) SetName(name string) {
	ld.Name = name
}

func (ld LinkDescription) CheckList(obj bson.ObjectID, name string) error {
	if ld.ObjectId != obj {
		return fmt.Errorf("Link check failed: Object Id mismatch")
	}

	if ld.Name != name {
		return fmt.Errorf("Link check failed: Name mismatch")
	}

	return nil
}

type UserToPageLink struct {
	kosmos.BaseModel `bson:",inline" kosmos:"user_page"`
	LinkDescription  `bson:",inline"`
}

func (l UserToPageLink) SelfScope() expression.Scope {
	return expression.CreateScope(
		kosmos.Fld("SubjectId").Eq(l.SubjectId),
		kosmos.Fld("ObjectId").Eq(l.ObjectId),
		kosmos.Fld("Name").Eq(l.Name),
	)
}

type UserToCommentLink struct {
	kosmos.BaseModel `bson:",inline" kosmos:"rel_commentrelation"`
	LinkDescription  `bson:",inline"`
}

func (l UserToCommentLink) SelfScope() expression.Scope {
	return expression.CreateScope(
		kosmos.Fld("SubjectId").Eq(l.SubjectId),
		kosmos.Fld("ObjectId").Eq(l.ObjectId),
		kosmos.Fld("Name").Eq(l.Name),
	)
}
