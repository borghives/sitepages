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

type SitePage struct {
	XMLName          xml.Name             `xml:"page"`
	ID               primitive.ObjectID   `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	Root             primitive.ObjectID   `xml:"data-root,attr" json:"Root" bson:"root"`
	Link             string               `xml:"Link,omitempty" json:"Link" bson:"link"`
	Title            string               `xml:"Title" json:"Title" bson:"title"`
	Abstract         string               `xml:"Abstract,omitempty" json:"Abstract,omitempty" bson:"abstract,omitempty"`
	Image            string               `xml:"Image,omitempty" json:"Image,omitempty" bson:"image,omitempty"`
	Synapses         []Synapse            `xml:"Synapse,omitempty" json:"Synapses,omitempty" bson:"synapses,omitempty"`
	Contents         []primitive.ObjectID `xml:"contents>content,omitempty" json:"Contents,omitempty" bson:"contents,omitempty"`
	Infos            MetaInfo             `xml:"_" json:"Infos,omitempty" bson:"infos,omitempty"`
	UpdatedTime      time.Time            `xml:"-" json:"UpdatedTime" bson:"updated_time"`
	PreviousVersion  primitive.ObjectID   `xml:"-" json:"PreviousVersion" bson:"previous_version"`
	CreatorSessionId primitive.ObjectID   `xml:"-" json:"-" bson:"session_id"`
	ContentData      []Stanza             `xml:"data>stanza,omitempty" json:"ContentData,omitempty" bson:"content_data,omitempty"`
}

type Stanza struct {
	XMLName            xml.Name           `xml:"stanza"`
	ID                 primitive.ObjectID `xml:"id,attr" json:"ID" bson:"_id,omitempty"`
	Root               primitive.ObjectID `xml:"data-root,attr" json:"Root" bson:"root"`
	Content            string             `xml:"Content" json:"Content" bson:"content"`
	UpdatedTime        time.Time          `xml:"-" json:"UpdatedTime" bson:"updated_time"`
	Context            primitive.ObjectID `xml:"-" json:"Context,omitempty" bson:"context,omitempty"`
	BasePage           primitive.ObjectID `xml:"data-base,attr" json:"BasePage" bson:"base_page"`
	PreviousVersion    primitive.ObjectID `xml:"data-prev,attr" json:"PreviousVersion" bson:"previous_version"`
	PreviousVersionIdx uint16             `xml:"data-previdx,attr" json:"PreviousVersionIdx" bson:"previous_version_idx"`
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
	Link string `json:"Link" bson:"link"`
	Name string `json:"Name" bson:"name"`
}

type MetaInfo struct {
	Contributors string     `json:"Contributors" bson:"contributors"`
	Promoters    string     `json:"Promoters" bson:"promoters"`
	Deeper       []LinkInfo `json:"Deeper" bson:"deeper"`
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
