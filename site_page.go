package sitepages

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SitePage struct {
	ID       primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Link     string             `json:"Link" bson:"link"`
	Title    string             `json:"Title" bson:"title"`
	SCards   []Illustrated      `json:"SCards" bson:"scards"`
	Contents []Stanza           `json:"Contents" bson:"content"`
	ECards   []Illustrated      `json:"ECards" bson:"ecards"`
	Infos    MetaInfo           `json:"Infos" bson:"infos"`
}

type Illustrated struct {
	ID      primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Title   string             `json:"Title" bson:"title"`
	Image   string             `json:"Image" bson:"image"`
	Content Stanza             `json:"Content" bson:"content"`
}

type Stanza struct {
	ID               primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Content          string             `json:"Content" bson:"content"`
	CreateTime       time.Time          `json:"CreateTime" bson:"create_time"`
	BasePage         primitive.ObjectID `json:"BasePage" bson:"base_page"`
	PreviousVersion  primitive.ObjectID `json:"PreviousVersion" bson:"previous_version"`
	InheritedVersion primitive.ObjectID `json:"InhertedVersion" bson:"inherted_version"`
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

func LoadSitePages(site string) map[string]SitePage {
	file, err := os.Open(site)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var retval map[string]SitePage
	err = json.NewDecoder(file).Decode(&retval)
	if err != nil {
		log.Fatal(err)
	}
	return retval
}
