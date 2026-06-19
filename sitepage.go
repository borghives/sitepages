package sitepages

import (
	"encoding/xml"
	"time"

	"git.mypierian.com/borghives/kosmos-go"
	"go.mongodb.org/mongo-driver/v2/bson"
)

//TODO: Concept of divergence (when different than that of most people agreement)
//a good concept can withstand the decay of Time
//Trust takes time to build

var MAX_LINK_LENGTH = 127
var MAX_TITLE_LENGTH = 255
var MAX_CHUNK_INDEX = 100
var MAX_ABSTRACT_LENGTH = 255

type Page struct {
	kosmos.BaseModel `bson:",inline" kosmos:"page"`
	XMLName          xml.Name        `xml:"page" json:"-" bson:"-"`
	Root             bson.ObjectID   `xml:"root" json:"root" bson:"root"`
	LinkName         string          `xml:"linkname,omitempty" json:"linkname" bson:"linkname"`
	Title            string          `xml:"title" json:"title" bson:"title"`
	Abstract         string          `xml:"abstract,omitempty" json:"abstract,omitempty" bson:"abstract,omitempty"`
	Synapses         []Synapse       `xml:"synapse,omitempty" json:"synapses,omitempty" bson:"synapses,omitempty"`
	Contents         []bson.ObjectID `xml:"contents>content,omitempty" json:"contents,omitempty" bson:"contents,omitempty"`
	Infos            MetaInfo        `xml:"infos,omitempty" json:"infos,omitzero" bson:"infos,omitempty"`
	Author           string          `xml:"author,omitempty" json:"author,omitempty" bson:"author,omitempty"`
	EventAt          time.Time       `xml:"eventat" json:"eventat" bson:"event_at"`
	PreviousVersion  bson.ObjectID   `xml:"previousversion" json:"previousversion" bson:"previous_version"`
	CreatorSessionID bson.ObjectID   `xml:"-" json:"-" bson:"session_id"`
	StanzaData       []Stanza        `xml:"-" json:"StanzaData,omitempty" bson:"stanza_data,omitempty"` //mainly for aggregate querying and not for storing into database or display as xml model
}

type Bundle struct {
	kosmos.BaseModel `bson:",inline" kosmos:"bundle"`
	XMLName          xml.Name        `xml:"bundle" json:"-" bson:"-"`
	Name             string          `xml:"name,omitempty" json:"name,omitempty" bson:"name,omitempty"`
	Contents         []bson.ObjectID `xml:"contents>content,omitempty" json:"contents,omitempty" bson:"contents,omitempty"`
	EventAt          time.Time       `xml:"eventat" json:"eventat" bson:"event_at"`
	PreviousBundleId bson.ObjectID   `xml:"previousbundleid" json:"previousbundleid" bson:"previous_bundle_id"`
	PageData         []Page          `xml:"pagedata" json:"PageData,omitempty" bson:"page_data,omitempty"`
}

type Stanza struct {
	kosmos.BaseModel `bson:",inline" kosmos:"stanza"`
	XMLName          xml.Name      `xml:"stanza" json:"-" bson:"-"`
	Content          string        `xml:"content" json:"content" bson:"content"`
	Context          bson.ObjectID `xml:"context,omitempty" json:"context,omitempty" bson:"context,omitempty"`
	BasePage         bson.ObjectID `xml:"basepage" json:"basepage" bson:"base_page"`
	PreviousVersion  bson.ObjectID `xml:"previousversion" json:"previousversion" bson:"previous_version"`
	TtlStart         time.Time     `xml:"-" json:"-" bson:"ttl_start,omitempty"`
}

type PageList struct {
	kosmos.BaseModel `bson:",inline" kosmos:"pagelist"`
	XMLName          xml.Name        `xml:"pagelist" json:"-" bson:"-"`
	Contents         []bson.ObjectID `xml:"contents>content,omitempty" json:"Contents,omitempty" bson:"contents,omitempty"`
	PageData         []Page          `xml:"-" json:"PageData,omitempty" bson:"page_data,omitempty"`
}

type Comment struct {
	kosmos.BaseModel `bson:",inline" kosmos:"comment"`
	XMLName          xml.Name      `xml:"comment" json:"-" bson:"-"`
	Root             bson.ObjectID `xml:"root" json:"root" bson:"root"`
	Parent           bson.ObjectID `xml:"parent" json:"parent" bson:"parent"`
	UserName         string        `xml:"username,omitempty" json:"username,omitempty" bson:"user_name,omitempty"`
	Moment           string        `xml:"moment" json:"moment" bson:"moment"`
	Content          string        `xml:"content" json:"content" bson:"content"`
	Infos            MetaInfo      `xml:"infos,omitempty" json:"Infos" bson:"infos,omitempty"`
	Score            float64       `xml:"-" json:"-" bson:"score"`
	BestBy           time.Time     `xml:"-" json:"-" bson:"best_by"`
	EventAt          time.Time     `xml:"eventat" json:"eventat" bson:"event_at"`
	TtlStart         time.Time     `xml:"-" json:"-" bson:"ttl_start,omitempty"`
}

type Synapse struct {
	FromPageId bson.ObjectID `json:"pageId" bson:"page_id"`
	FromStanza bson.ObjectID `json:"stanza" bson:"stanza"`
	ToPageId   bson.ObjectID `json:"topageid" bson:"to_page_id"`
	ToStanza   bson.ObjectID `json:"tostanza" bson:"to_stanza"`
	Info       string        `json:"info" bson:"info"`
}

type MetaInfo struct {
	SourceId        bson.ObjectID `xml:"sourceid,omitempty" json:"sourceid,omitempty" bson:"sourceid,omitempty"`
	Source          string        `xml:"source,omitempty" json:"source,omitempty" bson:"source,omitempty"`
	Category        string        `xml:"category,omitempty" json:"category,omitempty" bson:"category,omitempty"`
	HasMarketImpact bool          `xml:"hasmarketimpact,omitempty" json:"hasmarketimpact,omitempty" bson:"has_market_impact,omitempty"`
	GenType         string        `xml:"-" json:"gentype,omitempty" bson:"gen_type,omitempty"`
	Tags            []string      `xml:"tags,omitempty" json:"tags,omitempty" bson:"tags,omitempty"`
}
