package sitepages

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//TODO: Concept of divergence (when different than that of most people agreement)
//a good concept can withstand the decay of Time
//Trust takes time to build

var MAX_LINK_LENGTH = 127
var MAX_TITLE_LENGTH = 255
var MAX_CHUNK_INDEX = 100
var MAX_ABSTRACT_LENGTH = 255

type SitePage struct {
	XMLName          xml.Name             `xml:"page" json:"-" bson:"-"`
	ID               primitive.ObjectID   `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	Root             primitive.ObjectID   `xml:"root" json:"Root" bson:"root"`
	LinkName         string               `xml:"linkname,omitempty" json:"LinkName" bson:"link"`
	Title            string               `xml:"title" json:"Title" bson:"title"`
	Abstract         string               `xml:"abstract,omitempty" json:"Abstract,omitempty" bson:"abstract,omitempty"`
	Image            string               `xml:"image,omitempty" json:"Image,omitempty" bson:"image,omitempty"`
	Synapses         []Synapse            `xml:"synapse,omitempty" json:"Synapses,omitempty" bson:"synapses,omitempty"`
	Contents         []primitive.ObjectID `xml:"contents>content,omitempty" json:"Contents,omitempty" bson:"contents,omitempty"`
	Infos            MetaInfo             `xml:"infos,omitempty" json:"Infos,omitempty" bson:"infos,omitempty"`
	EventAt          time.Time            `xml:"eventat" json:"EventAt" bson:"event_at"`
	UpdatedTime      time.Time            `xml:"updated" json:"updated" bson:"updated_time"`
	PreviousVersion  primitive.ObjectID   `xml:"previousversion" json:"PreviousVersion" bson:"previous_version"`
	CreatorSessionId primitive.ObjectID   `xml:"-" json:"-" bson:"session_id"`
	StanzaData       []Stanza             `xml:"-" json:"StanzaData,omitempty" bson:"stanza_data,omitempty"` //mainly for aggregate querying and not for storing into database or display as xml model
}

type Bundle struct {
	XMLName          xml.Name             `xml:"bundle" json:"-" bson:"-"`
	ID               primitive.ObjectID   `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	Contents         []primitive.ObjectID `xml:"contents>content,omitempty" json:"Contents,omitempty" bson:"contents,omitempty"`
	EventAt          time.Time            `xml:"eventat" json:"EventAt" bson:"event_at"`
	PreviousBundleId primitive.ObjectID   `xml:"previousbundleid" json:"PreviousBundleId" bson:"previous_bundle_id"`
	PageData         []SitePage           `xml:"-" json:"PageData,omitempty" bson:"page_data,omitempty"` //for aggregate querying and not for storing into database or display as xml model
}

type Stanza struct {
	XMLName         xml.Name           `xml:"stanza" json:"-" bson:"-"`
	ID              primitive.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	Content         string             `xml:"content" json:"Content" bson:"content"`
	UpdatedTime     time.Time          `xml:"-" json:"UpdatedTime" bson:"updated_time"`
	Context         primitive.ObjectID `xml:"context,omitempty" json:"Context,omitempty" bson:"context,omitempty"`
	BasePage        primitive.ObjectID `xml:"basepage" json:"BasePage" bson:"base_page"`
	PreviousVersion primitive.ObjectID `xml:"previousversion" json:"PreviousVersion" bson:"previous_version"`
	ChunkIndex      uint16             `xml:"chunkidx" json:"ChunkIdx" bson:"-"`       //only for control and not for persist in state
	ChunkOffset     uint16             `xml:"chunkoffset" json:"ChunkOffset" bson:"-"` //only for control and not for persist in state
}

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
	XMLName  xml.Name           `xml:"relationship" json:"-" bson:"-"`
	SourceId primitive.ObjectID `xml:"-" json:"-" bson:"source_id"`
	TargetId primitive.ObjectID `xml:"targetid" json:"TargetId" bson:"target_id"`
	Root     primitive.ObjectID `xml:"root" json:"Root" bson:"root"`
	Relation RelationType       `xml:"relation" json:"Relation" bson:"relation"`
	Rank     float32            `xml:"rank" json:"Rank" bson:"rank"`
	EventAt  time.Time          `xml:"-" json:"-" bson:"event_at"`
	Type     RelationGraphType  `xml:"type,omitempty" json:"Type,omitempty" bson:"type,omitempty"`
	PageData []SitePage         `xml:"-,omitempty" json:"PageData,omitempty" bson:"page_data,omitempty"` //for aggregate querying and not for storing
}

type Comment struct {
	XMLName  xml.Name           `xml:"comment" json:"-" bson:"-"`
	ID       primitive.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	Root     primitive.ObjectID `xml:"root" json:"Root" bson:"root"`
	Parent   primitive.ObjectID `xml:"parent" json:"Parent" bson:"parent"`
	UserName string             `xml:"username,omitempty" json:"UserName,omitempty" bson:"user_name,omitempty"`
	Moment   string             `xml:"moment" json:"Moment" bson:"moment"`
	Content  string             `xml:"content" json:"Content" bson:"content"`
	Score    float32            `xml:"-" json:"-" bson:"score"`
	EventAt  time.Time          `xml:"eventat" json:"EventAt" bson:"event_at"`
}

type Synapse struct {
	FromPageId  primitive.ObjectID `json:"PageId" bson:"page_id"`
	FromStanza  primitive.ObjectID `json:"Stanza" bson:"stanza"`
	ToPageId    primitive.ObjectID `json:"ToPageId" bson:"to_page_id"`
	ToStanza    primitive.ObjectID `json:"ToStanza" bson:"to_stanza"`
	Info        string             `json:"Info" bson:"info"`
	UpdatedTime time.Time          `json:"UpdatedTime" bson:"updated_time"`
}

type LinkInfo struct {
	Link string `xml:"link,omitempty" json:"Link" bson:"link"`
	Name string `xml:"name,omitempty" json:"Name" bson:"name"`
}

type MetaInfo struct {
	Source string     `xml:"source,omitempty" json:"Source,omitempty" bson:"source,omitempty"`
	Deeper []LinkInfo `xml:"deeper,omitempty" json:"Deeper,omitempty" bson:"deeper,omitempty"`
}

type Princigo struct { //a deeper private self: an entity that mediates between our instincts and the social word.  It caries multiple persona, the social mask we wear.
	ID       primitive.ObjectID   `json:"ID" bson:"_id,omitempty"`
	Name     string               `json:"Name" bson:"name"`         //a public label of the self
	Personas []primitive.ObjectID `json:"Personas" bson:"personas"` //collection of personas that bridge the self to the world
	Shem     string               `bson:"shem"`                     //a private label to link with the world
}

type Persona struct {
	ID       primitive.ObjectID `json:"ID,omitempty" bson:"_id,omitempty"`
	EgoId    primitive.ObjectID `json:"EgoId,omitempty" bson:"_id,omitempty"`
	Labeling string             `json:"labeling,omitempty" bson:"labeling,omitempty"` //ex: "a Philosipher, a Scientist, and an Engineer"
}

type Illustrated struct {
	ID      primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Title   string             `json:"Title" bson:"title"`
	Image   string             `json:"Image" bson:"image"`
	Content Stanza             `json:"Content" bson:"content"`
}

func SaveSitePages(file string, pages []SitePage) error {
	// Open the file for writing
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(pages)
}

func GeneratePageToken(session WebSession, refPageId string, pageRootId string) string {
	return session.GenerateHexID("page" + refPageId + pageRootId)
}

func GenerateStanzaToken(session WebSession, pageId string, referenceStanza string, index uint16) string {
	return session.GenerateHexID("stanza" + pageId + referenceStanza + strconv.Itoa(int(index)))
}

func GenerateMomentString() string {
	return time.Now().UTC().Format("2006-01-02 15:04")
}

func ParseMomementString(moment string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04", moment)
}

func GenerateCommentToken(session WebSession, rootId string, salt string) string {
	return session.GenerateHexID("comment" + rootId + salt)
}

func GenerateRelationToken(session WebSession, rootId string, salt string) string {
	return session.GenerateHexID("relation" + rootId + salt)
}

func LoadSitePages(site string) []SitePage {
	file, err := os.Open(site)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var retval []SitePage
	err = json.NewDecoder(file).Decode(&retval)
	if err != nil {
		log.Fatal(err)
	}
	return retval
}
