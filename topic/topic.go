package topic

import (
	"fmt"
	"log"
	"time"

	"github.com/borghives/entanglement"
	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ##### Page  #####
type Page sitepages.Page

func (p Page) GetRootID() bson.ObjectID {
	return p.Root
}

func (p Page) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	frame = frame.CreateSubFrame("page_system")

	pageID := p.ID.Hex()

	frame.EntangleProperty("pageid", pageID)
	frame.EntangleProperty("rootid", p.Root.Hex())

	nextPageID := frame.GenerateCorrelation(pageID)
	correlation.AddCorrelation("page", pageID, nextPageID)
	correlation.Update(EntangleStanzaProperties(frame, nextPageID, p.Contents...))
	correlation.Update(EntangleCommentProperties(frame, p.ID, p.Root, 0))
	// log.Println("Entangle Page Id", nextPageID, frame.StateString())
	return correlation
}

func (p Page) CheckTransition(frame entanglement.Session) error {
	frame = frame.CreateSubFrame("page_system")
	frame.EntangleProperty("pageid", p.PreviousVersion.Hex())
	frame.EntangleProperty("rootid", p.Root.Hex())
	correlatedId := frame.GenerateCorrelation(p.PreviousVersion.Hex())
	if correlatedId != p.ID.Hex() {
		log.Printf("Mismatch page id: %s, expected %s := pageid: %s rootid: %s", p.ID.Hex(), correlatedId, p.PreviousVersion.Hex(), p.Root.Hex())
		return fmt.Errorf("Failed ID Expectation")
	}
	return nil
}

// ##### Stanza  #####
type Stanza sitepages.Stanza

func (s Stanza) GetRootID() bson.ObjectID {
	return s.BasePage
}

func (s Stanza) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	return correlation
}

func (s Stanza) CheckTransition(frame entanglement.Session) error {
	frame = frame.CreateSubFrame("stanza_system")
	frame.EntangleProperty("baseid", s.BasePage.Hex())
	correlatedId := frame.GenerateCorrelation(s.PreviousVersion.Hex())

	if correlatedId != s.ID.Hex() {
		log.Printf("Mismatch stanza id: %s, expected %s := %s", s.ID.Hex(), correlatedId, frame.StateString())
		return fmt.Errorf("Failed ID Expectation")
	}
	return nil
}

func EntangleStanzaProperties(frame entanglement.Session, baseid string, contents ...bson.ObjectID) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	if len(contents) > 0 {
		frame = frame.CreateSubFrame("stanza_system")
		frame.EntangleProperty("baseid", baseid)
		for _, content := range contents {
			stanzaID := content.Hex()
			nextStanzaID := frame.GenerateCorrelation(stanzaID)
			// log.Println("Entangle Stanza Id", stanzaID, nextStanzaID, frame.StateString())
			correlation.AddCorrelation("stanza", stanzaID, nextStanzaID)
		}
	}
	return correlation
}

// ##### Comment #####
type Comment sitepages.Comment

func (c Comment) GetRootID() bson.ObjectID {
	return c.Root
}

func (c Comment) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	return correlation
}

func (c Comment) CheckTransition(frame entanglement.Session) error {
	now := time.Now().UTC()
	tmoment, err := sitepages.ParseMomentString(c.Moment)
	if err != nil {
		return fmt.Errorf("Check Comment Transition: %v", err)
	}

	age := now.Sub(tmoment.UTC())
	if age.Minutes() > 30 {
		return fmt.Errorf("Check Comment Transition: stale moment %v", c.Moment)
	}
	frame = frame.CreateSubFrame("comment_system")
	frame.EntangleProperty("sourceid", c.Infos.SourceId.Hex())
	frame.EntangleProperty("rootid", c.Root.Hex())
	frame.EntangleProperty("moment", c.Moment)
	derivedHexID := frame.GenerateCorrelation("--page-comment-creator")
	if derivedHexID != c.ID.Hex() {
		return fmt.Errorf("Check Comment Transition: mismatch comment id (%s) expect (%s): %s", c.ID.Hex(), derivedHexID, frame.StateString())
	}
	return nil
}

func EntangleCommentProperties(frame entanglement.Session, sourceId bson.ObjectID, rootId bson.ObjectID, coolDown time.Duration) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	moment := sitepages.GenerateMomentString(coolDown)

	frame = frame.CreateSubFrame("comment_system")
	frame.EntangleProperty("sourceid", sourceId.Hex())
	frame.EntangleProperty("rootid", rootId.Hex())
	frame.EntangleProperty("moment", moment)
	nextId := frame.GenerateCorrelation("--page-comment-creator")
	// log.Println("Entangle Comment Id", nextId, frame.StateString())
	correlation.AddCorrelation("comment", "--page-comment-creator", nextId)
	correlation.AddCorrelation("comment", "moment", moment)
	return correlation
}
