package sitepages

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SitePage struct {
	ID               primitive.ObjectID   `json:"ID" bson:"_id,omitempty"`
	Link             string               `json:"Link" bson:"link"`
	Title            string               `json:"Title" bson:"title"`
	SCards           []primitive.ObjectID `json:"SCards" bson:"scards"`
	Contents         []primitive.ObjectID `json:"Contents" bson:"contents"`
	ECards           []primitive.ObjectID `json:"ECards" bson:"ecards"`
	Infos            MetaInfo             `json:"Infos" bson:"infos"`
	UpdatedTime      time.Time            `json:"UpdatedTime" bson:"updated_time"`
	PreviousVersion  primitive.ObjectID   `json:"PreviousVersion" bson:"previous_version"`
	CreatorSessionId primitive.ObjectID   `json:"SessionId" bson:"session_id"`
}

type SitePageAgg struct {
	ID               primitive.ObjectID   `json:"ID" bson:"_id,omitempty"`
	Link             string               `json:"Link" bson:"link"`
	Title            string               `json:"Title" bson:"title"`
	SCards           []primitive.ObjectID `json:"SCards" bson:"scards"`
	Contents         []primitive.ObjectID `json:"Contents" bson:"contents"`
	ECards           []primitive.ObjectID `json:"ECards" bson:"ecards"`
	Infos            MetaInfo             `json:"Infos" bson:"infos"`
	UpdatedTime      time.Time            `json:"UpdatedTime" bson:"updated_time"`
	PreviousVersion  primitive.ObjectID   `json:"PreviousVersion" bson:"previous_version"`
	CreatorSessionId primitive.ObjectID   `json:"SessionId" bson:"session_id"`
	ContentData      []Stanza             `json:"ContentData" bson:"content_data"`
}

type Illustrated struct {
	ID      primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Title   string             `json:"Title" bson:"title"`
	Image   string             `json:"Image" bson:"image"`
	Content Stanza             `json:"Content" bson:"content"`
}

type Stanza struct {
	ID              primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	Content         string             `json:"Content" bson:"content"`
	UpdatedTime     time.Time          `json:"UpdatedTime" bson:"updated_time"`
	BasePage        primitive.ObjectID `json:"BasePage" bson:"base_page"`
	PreviousVersion primitive.ObjectID `json:"PreviousVersion" bson:"previous_version"`
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
