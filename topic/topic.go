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

func (t Page) GetRootID() bson.ObjectID {
	return t.Root
}

func (t Page) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	frame = frame.CreateSubFrame("page_system")

	pageID := t.ID.Hex()

	frame.EntangleProperty("pageid", pageID)
	frame.EntangleProperty("rootid", t.Root.Hex())

	nextPageID := frame.GenerateCorrelation(pageID)
	correlation.AddCorrelation("page", pageID, nextPageID)
	correlation.Update(EntangleStanzaProperties(frame, nextPageID, t.Contents...))
	correlation.Update(EntangleCommentProperties(frame, t.ID, t.Root, 0))
	log.Println("Entangle Page Id", nextPageID, frame.StateString())
	return correlation
}

func (t Page) CheckTransition(frame entanglement.Session) error {
	frame = frame.CreateSubFrame("page_system")
	frame.EntangleProperty("pageid", t.PreviousVersion.Hex())
	frame.EntangleProperty("rootid", t.Root.Hex())
	correlatedId := frame.GenerateCorrelation(t.PreviousVersion.Hex())
	if correlatedId != t.ID.Hex() {
		log.Printf("Missmatch page id: %s, expected %s := pageid: %s rootid: %s", t.ID.Hex(), correlatedId, t.PreviousVersion.Hex(), t.Root.Hex())
		return fmt.Errorf("Failed ID Expectation")
	}
	return nil
}

// ##### Stanza  #####
type Stanza sitepages.Stanza

func (t Stanza) GetRootID() bson.ObjectID {
	return t.BasePage
}

func (t Stanza) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	return correlation
}

func (t Stanza) CheckTransition(frame entanglement.Session) error {
	frame = frame.CreateSubFrame("stanza_system")
	frame.EntangleProperty("baseid", t.BasePage.Hex())
	correlatedId := frame.GenerateCorrelation(t.PreviousVersion.Hex())

	if correlatedId != t.ID.Hex() {
		log.Printf("Missmatch stanza id: %s, expected %s := %s", t.ID.Hex(), correlatedId, frame.StateString())
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
			log.Println("Entangle Stanza Id", stanzaID, nextStanzaID, frame.StateString())
			correlation.AddCorrelation("stanza", stanzaID, nextStanzaID)
		}
	}
	return correlation
}

// ##### Comment #####
type Comment sitepages.Comment

func (t Comment) GetRootID() bson.ObjectID {
	return t.Root
}

func (t Comment) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	return correlation
}

func (t Comment) CheckTransition(frame entanglement.Session) error {
	now := time.Now().UTC()
	tmoment, err := sitepages.ParseMomementString(t.Moment)
	if err != nil {
		return fmt.Errorf("Check Comment Transistion: %v", err)
	}

	age := now.Sub(tmoment.UTC())
	if age.Minutes() > 30 {
		return fmt.Errorf("Check Comment Transistion: stale moment %v", t.Moment)
	}
	frame = frame.CreateSubFrame("comment_system")
	frame.EntangleProperty("sourceid", t.Infos.SourceId.Hex())
	frame.EntangleProperty("rootid", t.Root.Hex())
	frame.EntangleProperty("moment", t.Moment)
	derivedHexID := frame.GenerateCorrelation("--page-comment-creator")
	log.Println("Comment Check Id ", derivedHexID, frame.StateString())
	if derivedHexID != t.ID.Hex() {
		return fmt.Errorf("Check Comment Transistion: mismatch comment id (%s) expect (%s)", t.ID.Hex(), derivedHexID)
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
	log.Println("Entangle Comment Id", nextId, frame.StateString())
	correlation.AddCorrelation("comment", "--page-comment-creator", nextId)
	correlation.AddCorrelation("comment", "moment", moment)
	return correlation
}
