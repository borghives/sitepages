package topic

import (
	"net/http"

	"github.com/borghives/entanglement"
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

func SetupEntanglement(r *http.Request) entanglement.SystemFrame {
	return entanglement.Create(r.Header.Get("x-entanglement-nonce"), r.Header.Get("x-entanglement-token"))
}
