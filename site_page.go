package sitepages

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//TODO: Concept of divergence (when different than that of most people agreement)
//a good concept can withstand the decay of Time
//Trust takes time to build

type SitePage struct {
	ID               primitive.ObjectID   `json:"ID" bson:"_id,omitempty"`
	Root             primitive.ObjectID   `json:"Root" bson:"root"`
	Link             string               `json:"Link" bson:"link"`
	Title            string               `json:"Title" bson:"title"`
	Contents         []primitive.ObjectID `json:"Contents" bson:"contents"`
	Infos            MetaInfo             `json:"Infos" bson:"infos"`
	UpdatedTime      time.Time            `json:"UpdatedTime" bson:"updated_time"`
	PreviousVersion  primitive.ObjectID   `json:"PreviousVersion" bson:"previous_version"`
	CreatorSessionId primitive.ObjectID   `bson:"session_id"`
}

type Stanza struct {
	ID                 primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Root               primitive.ObjectID `json:"Root" bson:"root"`
	Content            string             `json:"Content" bson:"content"`
	UpdatedTime        time.Time          `json:"UpdatedTime" bson:"updated_time"`
	Context            primitive.ObjectID `json:"Context" bson:"context"`
	BasePage           primitive.ObjectID `json:"BasePage" bson:"base_page"`
	PreviousVersion    primitive.ObjectID `json:"PreviousVersion" bson:"previous_version"`
	PreviousVersionIdx uint16             `json:"PreviousVersionIdx" bson:"previous_version_idx"`
}

type Synapse struct {
	FromPageId  primitive.ObjectID `json:"PageId" bson:"page_id"`
	FromStanza  primitive.ObjectID `json:"Stanza" bson:"stanza"`
	ToPageId    primitive.ObjectID `json:"ToPageId" bson:"to_page_id"`
	ToStanza    primitive.ObjectID `json:"ToStanza" bson:"to_stanza"`
	Info        string             `json:"Info" bson:"info"`
	UpdatedTime time.Time          `json:"UpdatedTime" bson:"updated_time"`
}

type SitePageAgg struct {
	ID               primitive.ObjectID   `json:"ID" bson:"_id,omitempty"`
	Root             primitive.ObjectID   `json:"Root" bson:"root"`
	Link             string               `json:"Link" bson:"link"`
	Title            string               `json:"Title" bson:"title"`
	Contents         []primitive.ObjectID `json:"Contents" bson:"contents"`
	Infos            MetaInfo             `json:"Infos" bson:"infos"`
	UpdatedTime      time.Time            `json:"UpdatedTime" bson:"updated_time"`
	PreviousVersion  primitive.ObjectID   `json:"PreviousVersion" bson:"previous_version"`
	CreatorSessionId primitive.ObjectID   `json:"SessionId" bson:"session_id"`
	ContentData      []Stanza             `json:"ContentData" bson:"content_data"`
}

type LinkInfo struct {
	Link string `json:"Link" bson:"link"`
	Name string `json:"Name" bson:"name"`
}

type MetaInfo struct {
	Contributors string     `json:"Contributors" bson:"_id,contributors"`
	Promoters    string     `json:"Promoters" bson:"promoters"`
	Deeper       []LinkInfo `json:"Deeper" bson:"deeper"`
}

type Illustrated struct {
	ID      primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Title   string             `json:"Title" bson:"title"`
	Image   string             `json:"Image" bson:"image"`
	Content Stanza             `json:"Content" bson:"content"`
}

func SaveSitePages(file string, pages []SitePageAgg) error {
	// Open the file for writing
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(pages)
}

func GeneratePageToken(session *WebSession, refPageId string, pageRootId string) string {
	return session.GenerateHexID("page" + refPageId + pageRootId)
}

func GenerateStanzaToken(session *WebSession, pageId string, referenceStanza string, index uint16) string {
	return session.GenerateHexID("stanza" + pageId + referenceStanza + strconv.Itoa(int(index)))
}

func LoadSitePages(site string) []SitePageAgg {
	file, err := os.Open(site)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var retval []SitePageAgg
	err = json.NewDecoder(file).Decode(&retval)
	if err != nil {
		log.Fatal(err)
	}
	return retval
}
