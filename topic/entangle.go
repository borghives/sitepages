package topic

import (
	"log"
	"net/http"
	"time"

	"github.com/borghives/entanglement"
	"github.com/borghives/entanglement/concept"
	"github.com/borghives/sitepages"
	"github.com/borghives/websession"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CorrelationMap map[string]string

type EntangleProperties struct {
	Token        string                    `xml:"-" json:"Token" bson:"-" `
	Correlations map[string]CorrelationMap `xml:"-" json:"Correlations,omitempty" bson:"-" `
}

func NewResponse() Response {
	return &EntangledResponse{}
}

func (e *EntangleProperties) SetCorrelationProperties(typeName string, properties CorrelationMap) {
	if e.Correlations == nil {
		e.Correlations = make(map[string]CorrelationMap)
	}

	e.Correlations[typeName] = properties
}

type Entangleable interface {
	SystemFrame() string
}

type EntangledResponse struct {
	BaseResponse
	EntanglementState *EntangleProperties `xml:"-" json:"EntangleProperties,omitempty" bson:"-" `
}

func (e *EntangledResponse) Append(data any) bson.ObjectID {
	switch data := data.(type) {
	case *concept.Entanglement:
		if e.EntanglementState == nil {
			e.EntanglementState = &EntangleProperties{}
		}
		e.EntanglementState.Token = data.GenerateToken()
		e.EntangleFrame(data)
		return bson.ObjectID{}
	default:
		return e.BaseResponse.Append(data)
	}
}

func (e *EntangledResponse) EntangleFrame(entanglement *concept.Entanglement) {
	switch entanglement.Frame {
	case "page_system":
		var page *sitepages.SitePage
		for _, data := range e.PageData {
			if data.ID.Hex() == e.TargetID.Hex() {
				page = &data
				break
			}
		}
		e.EntanglementState = EntanglePage(e.EntanglementState, entanglement, page)
	}
}

func EntanglePage(state *EntangleProperties, entanglement *concept.Entanglement, page *sitepages.SitePage) *EntangleProperties {
	if page == nil || state == nil {
		return state
	}

	entanglement.SetProperty("pageid", page.ID.Hex())
	entanglement.SetProperty("rootid", page.Root.Hex())

	pageId := entanglement.GenerateCorrelation(page.ID.Hex())
	state.SetCorrelationProperties("page", CorrelationMap{
		page.ID.Hex(): pageId,
	})

	stanzaEntanglement := entanglement.CreatSubFrame("stanza_system")
	stanzaEntanglement.SetProperty("baseid", pageId)

	if len(page.Contents) > 0 {
		stanzaCorrelation := CorrelationMap{}
		for _, content := range page.Contents {
			stanzaCorrelation[content.Hex()] = stanzaEntanglement.GenerateCorrelation(content.Hex())
		}

		state.Correlations["stanza"] = stanzaCorrelation
	}

	return state
}

type ServeEntangled struct {
	ServePipe
	DoCheck bool
	Frame   string
}

func HandleEntangled(frame string, doCheck bool, chain ...Handler) *ServeEntangled {
	pipe := &ServeEntangled{
		ServePipe: *Handle(chain...),
		DoCheck:   doCheck,
		Frame:     frame,
	}
	return pipe
}

func (s *ServeEntangled) ServeTopic(response Response, r *http.Request) {
	entanglement, err := SetupEntanglement(r)
	if err != nil {
		response.SetOnError(err, http.StatusExpectationFailed)
		return
	}

	if s.Frame != "" {
		entanglement.SetFrame(s.Frame)
	}

	if s.DoCheck {
		if err := entanglement.CheckToken(); err != nil {
			response.SetOnError(err, http.StatusExpectationFailed)
			return
		}
	}

	s.ServePipe.ServeTopic(response, r)

	if entanglement != nil {
		response.Append(entanglement)
	}

}

func EntangleCommentProperties(entanglement *concept.Entanglement, sourceId bson.ObjectID, rootId bson.ObjectID, coolDown time.Duration) CorrelationMap {
	moment := sitepages.GenerateMomentString(coolDown)
	commentEntanglement := entanglement.CreatSubFrame("comment_system")
	commentEntanglement.SetProperty("sourceid", sourceId.Hex())
	commentEntanglement.SetProperty("rootid", rootId.Hex())
	commentEntanglement.SetProperty("moment", moment)
	log.Println("sourceid", sourceId.Hex(), "rootid", rootId.Hex(), "moment", moment)
	return CorrelationMap{
		"--page-comment-creator": commentEntanglement.GenerateCorrelation("--page-comment-creator"),
		"moment":                 moment,
	}
}

func SetupEntanglement(r *http.Request) (*concept.Entanglement, error) {
	session, err := websession.Manager().GetAndVerifySession(r)
	if err != nil {
		return nil, err
	}
	return entanglement.CreateEntanglementWithNonceAndToken(
		session,
		r.Header.Get("x-entanglement-nonce"),
		r.Header.Get("x-entanglement-token"),
	), nil
}
