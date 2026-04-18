package sitepages

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"os"
	"time"

	"github.com/borghives/kosmos-go"
	"go.mongodb.org/mongo-driver/v2/bson"
)

//TODO: Concept of divergence (when different than that of most people agreement)
//a good concept can withstand the decay of Time
//Trust takes time to build

var MAX_LINK_LENGTH = 127
var MAX_TITLE_LENGTH = 255
var MAX_CHUNK_INDEX = 100
var MAX_ABSTRACT_LENGTH = 255

type SitePage struct {
	kosmos.BaseModel `bson:",inline" kosmos:"page"`
	XMLName          xml.Name        `xml:"page" json:"-" bson:"-"`
	Root             bson.ObjectID   `xml:"root" json:"Root" bson:"root"`
	LinkName         string          `xml:"linkname,omitempty" json:"LinkName" bson:"linkname"`
	Title            string          `xml:"title" json:"Title" bson:"title"`
	Abstract         string          `xml:"abstract,omitempty" json:"Abstract,omitempty" bson:"abstract,omitempty"`
	Image            string          `xml:"image,omitempty" json:"Image,omitempty" bson:"image,omitempty"`
	Synapses         []Synapse       `xml:"synapse,omitempty" json:"Synapses,omitempty" bson:"synapses,omitempty"`
	Contents         []bson.ObjectID `xml:"contents>content,omitempty" json:"Contents,omitempty" bson:"contents,omitempty"`
	Infos            MetaInfo        `xml:"infos,omitempty" json:"Infos,omitzero" bson:"infos,omitempty"`
	Authg            string          `xml:"authg,omitempty" json:"Authg,omitempty" bson:"authg,omitempty"`
	CommentCount     uint32          `xml:"commentcount" json:"CommentCount" bson:"comment_count"`
	EventAt          time.Time       `xml:"eventat" json:"EventAt" bson:"event_at"`
	PreviousVersion  bson.ObjectID   `xml:"previousversion" json:"PreviousVersion" bson:"previous_version"`
	CreatorSessionId bson.ObjectID   `xml:"-" json:"-" bson:"session_id"`
	StanzaData       []Stanza        `xml:"-" json:"StanzaData,omitempty" bson:"stanza_data,omitempty"` //mainly for aggregate querying and not for storing into database or display as xml model
}

type Bundle struct {
	kosmos.BaseModel `bson:",inline" kosmos:"bundle"`
	XMLName          xml.Name        `xml:"bundle" json:"-" bson:"-"`
	Name             string          `xml:"name,omitempty" json:"name,omitempty" bson:"name,omitempty"`
	Contents         []bson.ObjectID `xml:"contents>content,omitempty" json:"Contents,omitempty" bson:"contents,omitempty"`
	EventAt          time.Time       `xml:"eventat" json:"EventAt" bson:"event_at"`
	PreviousBundleId bson.ObjectID   `xml:"previousbundleid" json:"PreviousBundleId" bson:"previous_bundle_id"`
	PageData         []SitePage      `xml:"-" json:"PageData,omitempty" bson:"page_data,omitempty"` //for aggregate querying and not for storing into database or display as xml model
}

type Stanza struct {
	kosmos.BaseModel `bson:",inline" kosmos:"stanza"`
	XMLName          xml.Name      `xml:"stanza" json:"-" bson:"-"`
	Content          string        `xml:"content" json:"Content" bson:"content"`
	UpdatedTime      time.Time     `xml:"-" json:"UpdatedTime" bson:"updated_time"`
	Context          bson.ObjectID `xml:"context,omitempty" json:"Context,omitempty" bson:"context,omitempty"`
	BasePage         bson.ObjectID `xml:"basepage" json:"BasePage" bson:"base_page"`
	PreviousVersion  bson.ObjectID `xml:"previousversion" json:"PreviousVersion" bson:"previous_version"`
	TtlStart         time.Time     `xml:"-" json:"-" bson:"ttl_start,omitempty"`
	ChunkIndex       uint16        `xml:"chunkidx" json:"ChunkIdx" bson:"-"`                               //only for control and not for persist in state
	ChunkOffset      uint16        `xml:"chunkoffset" json:"ChunkOffset" bson:"-"`                         //only for control and not for persist in state
	Chunkings        []uint16      `xml:"chunkings>content,omitempty" json:"chunkings,omitempty" bson:"-"` //only for control and not for persist in state
}

type PageList struct {
	kosmos.BaseModel `bson:",inline" kosmos:"pagelist"`
	XMLName          xml.Name        `xml:"pagelist" json:"-" bson:"-"`
	Contents         []bson.ObjectID `xml:"contents>content,omitempty" json:"Contents,omitempty" bson:"contents,omitempty"`
	PageData         []SitePage      `xml:"-" json:"PageData,omitempty" bson:"page_data,omitempty"`
}

type Comment struct {
	kosmos.BaseModel `bson:",inline" kosmos:"comment"`
	XMLName          xml.Name      `xml:"comment" json:"-" bson:"-"`
	Root             bson.ObjectID `xml:"root" json:"Root" bson:"root"`
	Parent           bson.ObjectID `xml:"parent" json:"Parent" bson:"parent"`
	UserName         string        `xml:"username,omitempty" json:"UserName,omitempty" bson:"user_name,omitempty"`
	Moment           string        `xml:"moment" json:"Moment" bson:"moment"`
	Content          string        `xml:"content" json:"Content" bson:"content"`
	Infos            MetaInfo      `xml:"infos,omitempty" json:"Infos" bson:"infos,omitempty"`
	Score            float64       `xml:"-" json:"-" bson:"score"`
	BestBy           time.Time     `xml:"-" json:"-" bson:"best_by"`
	EventAt          time.Time     `xml:"eventat" json:"EventAt" bson:"event_at"`
	TtlStart         time.Time     `xml:"-" json:"-" bson:"ttl_start,omitempty"`
}

type Synapse struct {
	FromPageId  bson.ObjectID `json:"PageId" bson:"page_id"`
	FromStanza  bson.ObjectID `json:"Stanza" bson:"stanza"`
	ToPageId    bson.ObjectID `json:"ToPageId" bson:"to_page_id"`
	ToStanza    bson.ObjectID `json:"ToStanza" bson:"to_stanza"`
	Info        string        `json:"Info" bson:"info"`
	UpdatedTime time.Time     `json:"UpdatedTime" bson:"updated_time"`
}

type LinkInfo struct {
	Link string `xml:"link,omitempty" json:"Link" bson:"link"`
	Name string `xml:"name,omitempty" json:"Name" bson:"name"`
}

type MetaInfo struct {
	SourceId        bson.ObjectID `xml:"sourceid,omitempty" json:"SourceId,omitempty" bson:"sourceid,omitempty"`
	Source          string        `xml:"source,omitempty" json:"Source,omitempty" bson:"source,omitempty"`
	Category        string        `xml:"category,omitempty" json:"Category,omitempty" bson:"category,omitempty"`
	HasMarketImpact bool          `xml:"hasmarketimpact,omitempty" json:"HasMarketImpact,omitempty" bson:"has_market_impact,omitempty"`
	GenType         string        `xml:"-" json:"GenType,omitempty" bson:"gen_type,omitempty"`
	Tags            []string      `xml:"tags,omitempty" json:"Tags,omitempty" bson:"tags,omitempty"`
	Deeper          []LinkInfo    `xml:"deeper,omitempty" json:"Deeper,omitempty" bson:"deeper,omitempty"`
}

func (t SitePage) GetRootID() bson.ObjectID {
	return t.Root
}

func (t SitePage) SystemFrame() string {
	return "page_system"
}

func (t Stanza) GetRootID() bson.ObjectID {
	return t.BasePage
}

func (t Stanza) SystemFrame() string {
	return "stanza_system"
}

func (t Comment) GetRootID() bson.ObjectID {
	return t.Root
}

func (t Comment) SystemFrame() string {
	return "comment_system"
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

func GenerateMomentString(coolDown time.Duration) string {
	now := time.Now().UTC()
	return now.Add(coolDown).Format("2006-01-02 15:04")
}

func ParseMomementString(moment string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04", moment)
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
