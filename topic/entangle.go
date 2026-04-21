package topic

import (
	"log"
	"net/http"

	"github.com/borghives/entanglement"
	"github.com/borghives/websession"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EntangledResponse struct {
	BaseResponse
	EntanglementState *entanglement.EntangleProperties `xml:"-" json:"EntangleProperties,omitempty" bson:"-" `
}

func NewResponse() Response {
	return &EntangledResponse{}
}

func (e *EntangledResponse) Append(data any) bson.ObjectID {
	switch data := data.(type) {
	case *entanglement.Session:
		e.EntangleFrame(*data)
	case entanglement.Session:
		e.EntangleFrame(data)
	default:
		return e.BaseResponse.Append(data)
	}
	return bson.ObjectID{}
}

func (e *EntangledResponse) EntangleFrame(frameSession entanglement.Session) {
	if e.EntanglementState == nil {
		e.EntanglementState = &entanglement.EntangleProperties{}
	}

	e.EntanglementState.Token = frameSession.GenerateToken()
	topic := e.GetTarget()

	entity, ok := topic.(entanglement.Correlatable)
	if ok {
		correlation := entity.TransitionStates(frameSession)
		e.EntanglementState.UpdateCorrelationProperties(correlation)
	}
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
	log.Printf("Serve Frame: %s", s.Frame)
	frame := SetupEntanglement(r)

	if s.Frame != "" {
		frame.SetFrame(s.Frame)
	}

	var session *websession.Session
	var err error
	if s.DoCheck {
		session, err = websession.Manager().GetAndVerifySession(r)
		if err != nil {
			response.SetOnError(err, http.StatusExpectationFailed)
			return
		}

		if err := frame.VerifyTokenAlignment(*session); err != nil {
			response.SetOnError(err, http.StatusExpectationFailed)
			return
		}
	}

	s.Handler.ServeTopic(response, r)

	response.Append(entanglement.EntangleSession(frame, *session))

}

func (h ServeEntangled) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	topicResponse := h.CreateResponse()

	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, h.BodyLimit)
	}

	h.ServeTopic(topicResponse, r)
	MarshalResponse(topicResponse, w)
}

func SetupEntanglement(r *http.Request) entanglement.SystemFrame {
	return entanglement.Create(r.Header.Get("x-entanglement-nonce"), r.Header.Get("x-entanglement-token"))
}
