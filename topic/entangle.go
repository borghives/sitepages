package topic

import (
	"net/http"
	"strings"

	"github.com/borghives/entanglement"
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

func (e *EntangleProperties) SetCorrelationProperties(name string, properties CorrelationMap) {
	if e.Correlations == nil {
		e.Correlations = make(map[string]CorrelationMap)
	}

	e.Correlations[name] = properties
}

func (e EntangleProperties) GetCorrelationProperties(name string) CorrelationMap {
	if e.Correlations == nil {
		return nil
	}
	return e.Correlations[name]
}

func (e *EntangleProperties) AddCorrelation(frame entanglement.Session, origin string) {
	frameName := frame.Frame
	frameName = strings.TrimSuffix(frameName, "_system")
	properties := e.GetCorrelationProperties(frameName)
	if properties == nil {
		properties = make(CorrelationMap)
	}

	derived := frame.GenerateCorrelation(origin)
	properties[origin] = derived

	e.SetCorrelationProperties(frameName, properties)

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
	case *entanglement.Session:
		e.EntangleFrame(*data)
		return bson.ObjectID{}
	case entanglement.Session:
		e.EntangleFrame(data)
		return bson.ObjectID{}
	default:
		return e.BaseResponse.Append(data)
	}
}

func (e *EntangledResponse) EntangleFrame(frameSession entanglement.Session) {
	if e.EntanglementState == nil {
		e.EntanglementState = &EntangleProperties{}
	}

	e.EntanglementState.Token = frameSession.GenerateToken()
	switch frameSession.Frame {
	case "page_system":
		var page *sitepages.SitePage
		for _, data := range e.PageData {
			if data.ID.Hex() == e.TargetID.Hex() {
				page = &data
				break
			}
		}
		e.EntanglementState = EntanglePage(e.EntanglementState, frameSession, page)
	}
}

func EntanglePage(state *EntangleProperties, pageframe entanglement.Session, page *sitepages.SitePage) *EntangleProperties {
	if page == nil || state == nil {
		return state
	}

	pageframe.EntangleProperty("pageid", page.ID.Hex())
	pageframe.EntangleProperty("rootid", page.Root.Hex())

	pageId := pageframe.GenerateCorrelation(page.ID.Hex())
	state.AddCorrelation(pageframe, page.ID.Hex())

	// state.SetCorrelationProperties("page", CorrelationMap{
	// 	page.ID.Hex(): pageId,
	// })

	stanzaframe := pageframe.CreateSubFrame("stanza_system")
	stanzaframe.EntangleProperty("baseid", pageId)

	if len(page.Contents) > 0 {
		// stanzaCorrelation := CorrelationMap{}
		for _, content := range page.Contents {
			// stanzaCorrelation[content.Hex()] = stanzaframe.GenerateCorrelation(content.Hex())
			state.AddCorrelation(stanzaframe, content.Hex())
		}

		// state.SetCorrelationProperties("stanza", stanzaCorrelation)
	}

	return state
}

type ServeEntangled struct {
	Handler        Handler
	CreateResponse ResponseFactory
	BodyLimit      int64
	DoCheck        bool
	Frame          string
}

func (s *ServeEntangled) SetBodyLimit(limit int64) *ServeEntangled {
	s.BodyLimit = limit
	return s
}

func HandleEntangled(frame string, doCheck bool, handler Handler) *ServeEntangled {
	pipe := &ServeEntangled{
		Handler:        handler,
		CreateResponse: NewResponse,
		DoCheck:        doCheck,
		Frame:          frame,
	}
	return pipe.SetBodyLimit(1048576)
}

func (s ServeEntangled) ServeTopic(response Response, r *http.Request) {
	web := SetupEntanglement(r)

	if s.Frame != "" {
		web.SetFrame(s.Frame)
	}

	var session *websession.Session
	var err error
	if s.DoCheck {
		session, err = websession.Manager().GetAndVerifySession(r)
		if err != nil {
			response.SetOnError(err, http.StatusExpectationFailed)
			return
		}

		if err := web.VerifyTokenAlignment(*session); err != nil {
			response.SetOnError(err, http.StatusExpectationFailed)
			return
		}
	}

	s.Handler.ServeTopic(response, r)

	response.Append(entanglement.EntangleSession(web, *session))

}

func (h ServeEntangled) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	topicResponse := h.CreateResponse()

	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, h.BodyLimit)
	}

	h.ServeTopic(topicResponse, r)
	MarshalResponse(topicResponse, w)
}

// func EntangleCommentProperties(web entanglement.Web, sourceId bson.ObjectID, rootId bson.ObjectID, coolDown time.Duration) CorrelationMap {
// 	moment := sitepages.GenerateMomentString(coolDown)
// 	commentEntanglement := web.CreateSubFrame("comment_system")
// 	commentEntanglement.SetProperty("sourceid", sourceId.Hex())
// 	commentEntanglement.SetProperty("rootid", rootId.Hex())
// 	commentEntanglement.SetProperty("moment", moment)
// 	log.Println("sourceid", sourceId.Hex(), "rootid", rootId.Hex(), "moment", moment)
// 	return CorrelationMap{
// 		"--page-comment-creator": commentEntanglement.GenerateCorrelation("--page-comment-creator"),
// 		"moment":                 moment,
// 	}
// }

func SetupEntanglement(r *http.Request) entanglement.SystemFrame {
	return entanglement.CreateWeb(r.Header.Get("x-entanglement-nonce"), r.Header.Get("x-entanglement-token"))
}
