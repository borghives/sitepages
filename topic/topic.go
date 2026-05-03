package topic

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/borghives/entanglement"
	"github.com/borghives/kosmos-go"
	"github.com/borghives/sitepages"
	"github.com/borghives/websession"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Hiarchical interface {
	GetRootID() bson.ObjectID
}

type Sanitizable interface {
	Sanitize(context RequestContext) error
}

// ##### Page  #####
type Page sitepages.Page

func (p Page) GetRootID() bson.ObjectID {
	return p.Root
}

func (p *Page) Sanitize(context RequestContext) error {

	if context.RootId != nil {
		p.Root = *context.RootId
		if context.HasUserName() && p.Root.IsZero() {
			//new page Root
			p.Root = bson.NewObjectID()
			p.Author = context.userSession.UserName
		}
	}

	if context.Request.Method == "PUT" {
		if !context.HasUserName() {
			return NewStatusString("Unauthorized to change page", http.StatusUnauthorized)
		}

		if p.Author != "" && p.Author != context.userSession.UserName {
			return NewStatusString("Unauthorized to change other user page", http.StatusUnauthorized)
		}

		if p.Root.IsZero() {
			return NewStatusString("Page root is zero", http.StatusBadRequest)
		}

		p.CreatorSessionID = context.userSession.ID
		p.LinkName = websession.MakeUniqueURL(p.Title, p.CreatorSessionID[:], p.Root.Hex(), p.Author)
	}

	return nil
}

func (p Page) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	frame = frame.CreateSubFrame("page_system")

	pageID := p.ID.Hex()

	frame.EntangleProperty("pageid", pageID)
	frame.EntangleProperty("rootid", p.Root.Hex())

	nextPageID := frame.GenerateCorrelation(pageID)
	correlation.AddCorrelation("page", pageID, nextPageID)
	correlation.Update(EntangleStanzaProperties(frame, nextPageID, 0, p.Contents...))
	correlation.Update(EntangleCommentProperties(frame, p.ID, p.Root, 0))
	slog.Debug("Entangle Page Id",
		slog.String("NextId", nextPageID),
		slog.String("FrameState", frame.StateString()),
	)
	return correlation
}

func (p Page) CheckTransition(frame entanglement.Session) error {
	frame = frame.CreateSubFrame("page_system")
	frame.EntangleProperty("pageid", p.PreviousVersion.Hex())
	frame.EntangleProperty("rootid", p.Root.Hex())
	correlatedId := frame.GenerateCorrelation(p.PreviousVersion.Hex())
	if correlatedId != p.ID.Hex() {
		log.Printf("Mismatch page id: %s, expected %s := pageid: %s rootid: %s", p.ID.Hex(), correlatedId, p.PreviousVersion.Hex(), p.Root.Hex())
		return NewStatusString("Failed ID Expectation", http.StatusExpectationFailed)
	}
	return nil
}

// ##### Stanza  #####
type Stanza struct {
	sitepages.Stanza `bson:",inline"`
	ChunkIndex       int   `xml:"chunkidx" json:"ChunkIdx" bson:"-"`
	ChunkOffset      int   `xml:"chunkoffset" json:"ChunkOffset" bson:"chunkoffset"`
	Chunkings        []int `xml:"chunkings>content,omitempty" json:"chunkings,omitempty" bson:"-"`
}

func (s Stanza) GetRootID() bson.ObjectID {
	return s.BasePage
}

func (s Stanza) TransitionStates(frame entanglement.Session) entanglement.TypeStateCorrelation {
	return EntangleStanzaProperties(frame, s.BasePage.Hex(), s.ChunkIndex, s.ID)
}

func (s Stanza) CheckTransition(frame entanglement.Session) error {
	frame = frame.CreateSubFrame("stanza_system")
	frame.EntangleProperty("baseid", s.BasePage.Hex())
	frame.EntangleProperty("chunkidx", strconv.Itoa(s.ChunkIndex))
	correlatedId := frame.GenerateCorrelation(s.PreviousVersion.Hex())

	if correlatedId != s.ID.Hex() {
		slog.Debug("Mismatch stanza id",
			slog.String("ID", s.ID.Hex()),
			slog.String("ExpectedID", correlatedId),
			slog.String("FrameState", frame.StateString()),
		)
		return NewStatusString("Failed ID Expectation", http.StatusExpectationFailed)
	}
	return nil
}

func CreateEntangledStanza(session entanglement.Session, content string, prevId bson.ObjectID, baseid bson.ObjectID, index int, finalOffset int) Stanza {
	session = session.CreateSubFrame("stanza_system")
	session.EntangleProperty("baseid", baseid.Hex())
	session.EntangleProperty("chunkidx", strconv.Itoa(index))
	newIdStr := session.GenerateCorrelation(prevId.Hex())
	newId, err := bson.ObjectIDFromHex(newIdStr)
	if err != nil {
		slog.Error("Error CreateStanza: newId", "error", err)
		return Stanza{}
	}

	return Stanza{
		Stanza: sitepages.Stanza{
			BaseModel: kosmos.BaseModel{
				ID: newId,
			},
			Content:         content,
			BasePage:        baseid,
			PreviousVersion: prevId,
		},
		ChunkIndex:  index,
		ChunkOffset: finalOffset,
	}
}

func SplitStanzaToOutput() HandlerFunc[Stanza] {
	return func(s *Session[Stanza]) error {
		session, err := s.GetVerifyEntanglement()
		if err != nil {
			return err
		}

		chunks := s.InBody.Chunkings
		slog.Debug("Chunking", slog.Any("chunks", chunks))
		switch len(chunks) {
		case 0:
			s.Output = append(s.Output, s.InBody)
		case 1:
			start := chunks[0]
			finalOffset := s.InBody.ChunkOffset + 2
			s.Output = append(s.Output,
				CreateEntangledStanza(*session, s.InBody.Content[:start], s.InBody.PreviousVersion, s.InBody.BasePage, s.InBody.ChunkOffset+1, finalOffset),
				CreateEntangledStanza(*session, s.InBody.Content[start:], s.InBody.PreviousVersion, s.InBody.BasePage, s.InBody.ChunkOffset+2, finalOffset),
			)
		case 2:
			start := chunks[0]
			end := chunks[1]
			finalOffset := s.InBody.ChunkOffset + 3
			if end > start {
				s.Output = append(s.Output,
					CreateEntangledStanza(*session, s.InBody.Content[:start], s.InBody.PreviousVersion, s.InBody.BasePage, s.InBody.ChunkOffset+1, finalOffset),
					CreateEntangledStanza(*session, s.InBody.Content[start:end], s.InBody.PreviousVersion, s.InBody.BasePage, s.InBody.ChunkOffset+2, finalOffset),
					CreateEntangledStanza(*session, s.InBody.Content[end:], s.InBody.PreviousVersion, s.InBody.BasePage, s.InBody.ChunkOffset+3, finalOffset),
				)
			}
		}

		return nil
	}
}

func EntangleStanzaProperties(frame entanglement.Session, baseid string, index int, contents ...bson.ObjectID) entanglement.TypeStateCorrelation {
	correlation := make(entanglement.TypeStateCorrelation)
	if len(contents) > 0 {
		frame = frame.CreateSubFrame("stanza_system")
		frame.EntangleProperty("baseid", baseid)
		frame.EntangleProperty("chunkidx", strconv.Itoa(index))
		for _, content := range contents {
			stanzaID := content.Hex()
			nextStanzaID := frame.GenerateCorrelation(stanzaID)
			slog.Debug("Entangle Stanza ID",
				slog.String("ID", stanzaID),
				slog.String("NextID", nextStanzaID),
				slog.String("FrameState", frame.StateString()),
			)
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
		slog.Debug("Check Comment Transition: mismatch comment id",
			slog.String("ID", c.ID.Hex()),
			slog.String("ExpectedID", derivedHexID),
			slog.String("FrameState", frame.StateString()),
		)
		return NewStatusString("Failed ID Expectation", http.StatusExpectationFailed)
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
