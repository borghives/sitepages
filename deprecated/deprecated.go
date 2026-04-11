package sitepages

import (
	"encoding/xml"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RelationType string

const (
	RelationType_Generic    RelationType = "generic"
	RelationType_Bookmarked RelationType = "bookmarked"
	RelationType_Endorsed   RelationType = "endorsed"
	RelationType_Objected   RelationType = "objected"
	RelationType_Ignored    RelationType = "ignored"
)

func (r RelationType) String() string {
	return string(r)
}

func CastRelationType(s string) RelationType {
	switch s {
	case "bookmarked", "endorsed", "objected", "ignored":
		return RelationType(s)
	default:
		return RelationType("generic")
	}
}

type RelationGraphType string

const (
	RelationGraphType_Opaque      RelationGraphType = "opaquerelation"
	RelationGraphType_UserPage    RelationGraphType = "pagerelation"
	RelationGraphType_UserComment RelationGraphType = "commentrelation"
)

func CastRelationGraphType(s string) RelationGraphType {
	switch s {
	case RelationGraphType_UserPage.String():
		return RelationGraphType_UserPage
	case RelationGraphType_UserComment.String():
		return RelationGraphType_UserComment
	default:
		return RelationGraphType_Opaque
	}
}

func (r RelationGraphType) String() string {
	return string(r)
}

type Relationship struct {
	XMLName  xml.Name          `xml:"relationship" json:"-" bson:"-"`
	SourceId bson.ObjectID     `xml:"-" json:"-" bson:"source_id"`
	TargetId bson.ObjectID     `xml:"targetid" json:"TargetId" bson:"target_id"`
	Root     bson.ObjectID     `xml:"root" json:"Root" bson:"root"`
	Relation RelationType      `xml:"relation" json:"Relation" bson:"relation"`
	Rank     float32           `xml:"rank" json:"Rank" bson:"rank"`
	EventAt  time.Time         `xml:"-" json:"-" bson:"event_at"`
	Type     RelationGraphType `xml:"type,omitempty" json:"Type,omitempty" bson:"type,omitempty"`
	PageData []SitePage        `xml:"-,omitempty" json:"PageData,omitempty" bson:"page_data,omitempty"` //for aggregate querying and not for storing
}

type Illustrated struct {
	ID      bson.ObjectID `json:"ID" bson:"_id,omitempty"`
	Title   string        `json:"Title" bson:"title"`
	Image   string        `json:"Image" bson:"image"`
	Content Stanza        `json:"Content" bson:"content"`
}

type Princigo struct { //a deeper private self: an entity that mediates between our instincts and the social word.  It caries multiple persona, the social mask we wear.
	ID       bson.ObjectID   `json:"ID" bson:"_id,omitempty"`
	Name     string          `json:"Name" bson:"name"`         //a public label of the self
	Personas []bson.ObjectID `json:"Personas" bson:"personas"` //collection of personas that bridge the self to the world
	Shem     string          `bson:"shem"`                     //a private label to link with the world
}

type Persona struct {
	ID       bson.ObjectID `json:"ID,omitempty" bson:"_id,omitempty"`
	EgoId    bson.ObjectID `json:"EgoId,omitempty" bson:"_id,omitempty"`
	Labeling string        `json:"labeling,omitempty" bson:"labeling,omitempty"` //ex: "a Philosipher, a Scientist, and an Engineer"
}
