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

func (br *BaseResponse) SetTargetID(id bson.ObjectID) {
	br.TargetID = id
}

func (br BaseResponse) GetTargetID() bson.ObjectID {
	return br.TargetID
}

func (br BaseResponse) GetTarget() Topic {
	if br.TargetID.IsZero() {
		return nil
	}

	for _, entity := range br.PageData {
		if entity.GetID() == br.TargetID {
			return entity
		}
	}

	for _, entity := range br.StanzaData {
		if entity.GetID() == br.TargetID {
			return entity
		}
	}

	for _, entity := range br.CommentData {
		if entity.GetID() == br.TargetID {
			return entity
		}
	}

	for _, entity := range br.BundleData {
		if entity.GetID() == br.TargetID {
			return entity
		}
	}

	return nil
}

func (br *BaseResponse) Append(data any) bson.ObjectID {
	var id bson.ObjectID
	switch response := data.(type) {
	case Page:
		br.PageData = append(br.PageData, response)
		id = response.ID
	case Stanza:
		br.StanzaData = append(br.StanzaData, response)
		id = response.ID
	case Comment:
		br.CommentData = append(br.CommentData, response)
		id = response.ID
	case sitepages.Bundle:
		br.BundleData = append(br.BundleData, response)
		id = response.ID
	case *Page:
		br.PageData = append(br.PageData, *response)
		id = response.ID
	case *Stanza:
		br.StanzaData = append(br.StanzaData, *response)
		id = response.ID
	case *Comment:
		br.CommentData = append(br.CommentData, *response)
		id = response.ID
	case *sitepages.Bundle:
		br.BundleData = append(br.BundleData, *response)
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
