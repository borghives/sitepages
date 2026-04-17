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
	StatusCode int    `xml:"-" json:"-" bson:"-" `
	StatusMsg  string `xml:"-" json:"error,omitempty" bson:"-" `
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
	TargetId    *bson.ObjectID       `xml:"-" json:"TargetId,omitempty" bson:"targetid,omitempty" `
	PageData    []sitepages.SitePage `xml:"-" json:"PageData,omitempty" bson:"pagedata,omitempty" `
	StanzaData  []sitepages.Stanza   `xml:"-" json:"StanzaData,omitempty" bson:"stanzadata,omitempty" `
	CommentData []sitepages.Comment  `xml:"-" json:"CommentData,omitempty" bson:"commentdata,omitempty" `
}

func NewBaseResponse() Response {
	return &BaseResponse{}
}

func (t *BaseResponse) SetTargetID(id bson.ObjectID) {
	t.TargetId = &id
}

func (t *BaseResponse) GetTargetID() *bson.ObjectID {
	return t.TargetId
}

func (t *BaseResponse) Append(data any) bson.ObjectID {
	var id bson.ObjectID
	switch response := data.(type) {
	case *sitepages.SitePage:
		t.PageData = append(t.PageData, *response)
		id = response.ID
	case *sitepages.Stanza:
		t.StanzaData = append(t.StanzaData, *response)
		id = response.ID
	case *sitepages.Comment:
		t.CommentData = append(t.CommentData, *response)
		id = response.ID
	case sitepages.SitePage:
		t.PageData = append(t.PageData, response)
		id = response.ID
	case sitepages.Stanza:
		t.StanzaData = append(t.StanzaData, response)
		id = response.ID
	case sitepages.Comment:
		t.CommentData = append(t.CommentData, response)
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
