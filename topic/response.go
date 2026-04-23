package topic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Topic interface {
	GetID() bson.ObjectID
}

type Response interface {
	SetTargetID(id bson.ObjectID)
	GetTargetID() bson.ObjectID
	GetTarget() Topic
	Append(data any) bson.ObjectID
}

type ErrorResponse interface {
	Error() string
	ErrorCode() int
}

type StatusResponse struct {
	StatusCode int    `json:"-" `
	StatusMsg  string `json:"message,omitempty" `
}

func NewStatusError(err error, code int) ErrorResponse {
	return &StatusResponse{StatusCode: code, StatusMsg: err.Error()}
}

func (e StatusResponse) GetStatus() StatusResponse {
	return e
}

func (e StatusResponse) ErrorCode() int {
	if e.StatusCode == 0 {
		return 500
	}
	return e.StatusCode
}

func (e StatusResponse) Error() string {
	return fmt.Sprintf("Response Status %d: %s", e.StatusCode, e.StatusMsg)
}

func (e StatusResponse) HasError() bool {
	return e.StatusCode >= 400
}

type BaseResponse struct {
	TargetID    bson.ObjectID      `json:"TargetId,omitempty,omitzero" `
	PageData    []Page             `json:"PageData,omitempty" `
	StanzaData  []Stanza           `json:"StanzaData,omitempty" `
	CommentData []Comment          `json:"CommentData,omitempty" `
	BundleData  []sitepages.Bundle `json:"BundleData,omitempty" `
}

func (t *BaseResponse) SetTargetID(id bson.ObjectID) {
	t.TargetID = id
}

func (t BaseResponse) GetTargetID() bson.ObjectID {
	return t.TargetID
}

func (t BaseResponse) GetTarget() Topic {
	if t.TargetID.IsZero() {
		return nil
	}

	for _, entity := range t.PageData {
		if entity.GetID() == t.TargetID {
			return entity
		}
	}

	for _, entity := range t.StanzaData {
		if entity.GetID() == t.TargetID {
			return entity
		}
	}

	for _, entity := range t.CommentData {
		if entity.GetID() == t.TargetID {
			return entity
		}
	}

	for _, entity := range t.BundleData {
		if entity.GetID() == t.TargetID {
			return entity
		}
	}

	return nil
}

func (t *BaseResponse) Append(data any) bson.ObjectID {
	var id bson.ObjectID
	switch response := data.(type) {
	case Page:
		t.PageData = append(t.PageData, response)
		id = response.ID
	case Stanza:
		t.StanzaData = append(t.StanzaData, response)
		id = response.ID
	case Comment:
		t.CommentData = append(t.CommentData, response)
		id = response.ID
	case sitepages.Bundle:
		t.BundleData = append(t.BundleData, response)
		id = response.ID
	case *Page:
		t.PageData = append(t.PageData, *response)
		id = response.ID
	case *Stanza:
		t.StanzaData = append(t.StanzaData, *response)
		id = response.ID
	case *Comment:
		t.CommentData = append(t.CommentData, *response)
		id = response.ID
	case *sitepages.Bundle:
		t.BundleData = append(t.BundleData, *response)
		id = response.ID
	default:
		log.Printf("Append unknown type: %T", data)
	}

	return id
}

func MarshalResponse[T Response](entity T, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)

}
