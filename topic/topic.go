package topic

import (
	"fmt"
	"log"

	"github.com/borghives/entanglement"
	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ##### Page  #####
type Page sitepages.Page

func (t Page) GetRootID() bson.ObjectID {
	return t.Root
}

func (p Page) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	frame.SetFrame("page_system")

	pageID := p.ID.Hex()

	frame.EntangleProperty("pageid", pageID)
	frame.EntangleProperty("rootid", p.Root.Hex())

	nextPageID := frame.GenerateCorrelation(pageID)
	correlation.AddCorrelation("page", pageID, nextPageID)

	stanzaframe := frame.CreateSubFrame("stanza_system")
	stanzaframe.EntangleProperty("baseid", nextPageID)

	if len(p.Contents) > 0 {
		for _, content := range p.Contents {
			stanzaID := content.Hex()
			nextStanzaID := stanzaframe.GenerateCorrelation(stanzaID)
			log.Printf("stanza expected %s := baseid: %s ", nextStanzaID, nextPageID)
			correlation.AddCorrelation("stanza", stanzaID, nextStanzaID)
		}
	}

	return correlation
}

func (p Page) CheckTransition(frame entanglement.Session) error {
	frame.SetFrame("page_system")
	frame.EntangleProperty("pageid", p.PreviousVersion.Hex())
	frame.EntangleProperty("rootid", p.Root.Hex())
	correlatedId := frame.GenerateCorrelation(p.PreviousVersion.Hex())
	if correlatedId != p.ID.Hex() {
		log.Printf("Missmatch page id: %s, expected %s := pageid: %s rootid: %s", p.ID.Hex(), correlatedId, p.PreviousVersion.Hex(), p.Root.Hex())
		return fmt.Errorf("Failed ID Expectation")
	}
	return nil
}

// ##### Stanza  #####
type Stanza sitepages.Stanza

func (t Stanza) GetRootID() bson.ObjectID {
	return t.BasePage
}

func (p Stanza) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	return correlation
}

func (s Stanza) CheckTransition(frame entanglement.Session) error {
	frame.SetFrame("stanza_system")
	frame.EntangleProperty("baseid", s.BasePage.Hex())
	correlatedId := frame.GenerateCorrelation(s.PreviousVersion.Hex())

	if correlatedId != s.ID.Hex() {
		log.Printf("Missmatch stanza id: %s, expected %s := baseid: %s index: %d", s.ID.Hex(), correlatedId, s.BasePage.Hex(), int(s.ChunkIndex))
		return fmt.Errorf("Failed ID Expectation")
	}
	return nil
}

// ##### Comment #####
type Comment sitepages.Comment

func (t Comment) GetRootID() bson.ObjectID {
	return t.Root
}
