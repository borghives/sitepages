package topic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Response interface {
	SetOnError(err error, code int) error
	GetStatus() StatusResponse
	HasError() bool
	SetTargetID(id bson.ObjectID)
	GetTargetID() *bson.ObjectID
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

func (e *StatusResponse) SetOnError(err error, code int) error {
	if err != nil {
		log.Printf("Response Error: %s", err.Error())
		if !e.HasError() {

			//Personal rule: only set the first error
			e.StatusMsg = err.Error()
			e.StatusCode = code
		}

		return e
	}
	return nil
}

func (e *StatusResponse) HasError() bool {
	return e.StatusCode >= 400
}

type BaseResponse struct {
	StatusResponse
	TargetID    *bson.ObjectID       `json:"TargetId,omitempty" `
	PageData    []sitepages.SitePage `json:"PageData,omitempty" `
	StanzaData  []sitepages.Stanza   `json:"StanzaData,omitempty" `
	CommentData []sitepages.Comment  `json:"CommentData,omitempty" `
	BundleData  []sitepages.Bundle   `json:"BundleData,omitempty" `
}

func (t *BaseResponse) SetTargetID(id bson.ObjectID) {
	t.TargetID = &id
}

func (t *BaseResponse) GetTargetID() *bson.ObjectID {
	return t.TargetID
}

func (t *BaseResponse) Append(data any) bson.ObjectID {
	var id bson.ObjectID
	switch response := data.(type) {
	case sitepages.SitePage:
		t.PageData = append(t.PageData, response)
		id = response.ID
	case sitepages.Stanza:
		t.StanzaData = append(t.StanzaData, response)
		id = response.ID
	case sitepages.Comment:
		t.CommentData = append(t.CommentData, response)
		id = response.ID
	case sitepages.Bundle:
		t.BundleData = append(t.BundleData, response)
		id = response.ID
	case *sitepages.SitePage:
		t.PageData = append(t.PageData, *response)
		id = response.ID
	case *sitepages.Stanza:
		t.StanzaData = append(t.StanzaData, *response)
		id = response.ID
	case *sitepages.Comment:
		t.CommentData = append(t.CommentData, *response)
		id = response.ID
	case *sitepages.Bundle:
		t.BundleData = append(t.BundleData, *response)
		id = response.ID
	}

	return id
}

func MarshalResponse[T Response](topic T, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if topic.HasError() {
		status := topic.GetStatus()
		w.WriteHeader(status.StatusCode)
		json.NewEncoder(w).Encode(status)
		return
	}

	json.NewEncoder(w).Encode(topic)

}
