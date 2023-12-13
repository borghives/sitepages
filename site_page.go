package sitepages

import (
	"encoding/json"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SitePage struct {
	ID       primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Link     string             `json:"Link" bson:"link"`
	Title    string             `json:"Title" bson:"title"`
	SCards   []Stanza           `json:"SCards" bson:"scards"`
	Contents []Stanza           `json:"Contents" bson:"content"`
	ECards   []Stanza           `json:"ECards" bson:"ecards"`
	Infos    MetaInfo           `json:"Infos" bson:"infos"`
}

type Stanza struct {
	ID      primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Link    string             `json:"Link" bson:"link"`
	Header  string             `json:"Header" bson:"header"`
	Image   string             `json:"Image" bson:"image"`
	Content string             `json:"Content" bson:"content"`
	Infos   MetaInfo           `json:"Infos" bson:"infos"`
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
