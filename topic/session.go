package topic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/borghives/entanglement"
	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/observation"
	"github.com/borghives/websession"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RequestSession struct {
	Request        *http.Request
	Response       Response
	EntangleFrame  entanglement.SystemFrame
	TopicId        *bson.ObjectID
	LatestTopic    bool
	urlQuery       *url.Values
	userSession    *websession.Session
	userSessionErr error
}

func NewRequestSession(r *http.Request) *RequestSession {
	return &RequestSession{
		Request: r,
		EntangleFrame: entanglement.Create(
			r.Header.Get("x-entanglement-nonce"),
			r.Header.Get("x-entanglement-token"),
		),
	}
}

func (t *RequestSession) URLQuery() url.Values {
	if t.urlQuery == nil {
		q := t.Request.URL.Query()
		t.urlQuery = &q
	}
	return *t.urlQuery
}

func (t *RequestSession) VerifySession() (*websession.Session, error) {
	if t.userSession == nil && t.userSessionErr == nil {
		t.userSession, t.userSessionErr = websession.Manager().GetAndVerifySession(t.Request)
	}

	return t.userSession, t.userSessionErr
}

type Session[T observation.Detectable] struct {
	RequestSession
	Detector *observation.EntityDetector[T]
	Body     T
}

func NewRequestTopicSession[T observation.Detectable](r *http.Request) *Session[T] {
	return &Session[T]{
		RequestSession: *NewRequestSession(r),
	}
}

func (t *Session[T]) TopicDetector() *observation.EntityDetector[T] {
	if t.Detector == nil {
		t.Detector = kosmos.All[T]()
	}

	return t.Detector
}

func (t *Session[T]) DecodeBody() error {
	if t.Request.Body == nil {
		return fmt.Errorf("Null Body in request")
	}

	return json.NewDecoder(t.Request.Body).Decode(&t.Body)
}

type HandlerFunc[T observation.Detectable] func(session *Session[T]) error
type Handler[T observation.Detectable] struct {
	Pipe []HandlerFunc[T]
}

func NewQuery[T observation.Detectable](responseCreator HandlerFunc[T]) *Handler[T] {
	session := &Handler[T]{}
	return session.Chain(responseCreator)
}

func (t *Handler[T]) Chain(chains ...HandlerFunc[T]) *Handler[T] {
	t.Pipe = append(t.Pipe, chains...)
	return t
}

func (t *Handler[T]) ChainTop(chain HandlerFunc[T]) *Handler[T] {
	t.Pipe = append([]HandlerFunc[T]{chain}, t.Pipe...)
	return t
}

func (t Handler[T]) AggregateSession(r *http.Request) (*Session[T], error) {
	session := NewRequestTopicSession[T](r)
	for _, chainExecution := range t.Pipe {
		if err := chainExecution(session); err != nil {
			return nil, err
		}
	}
	return session, nil
}

func ServeError(w http.ResponseWriter, err error) {
	log.Printf("Error Handling Topic Request Chain: %v", err)
	w.Header().Set("Content-Type", "application/json")
	status, ok := err.(ErrorResponse)
	if ok {
		w.WriteHeader(status.ErrorCode())
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func CreateEntangleResponse[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		if s.Response == nil {
			s.Response = NewResponse()
		}
		return nil
	}
}

func CreateRelationResponse[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		if s.Response == nil {
			s.Response = NewRelationTopicResponse()
		}
		return nil
	}
}

func CreateListResponse[T observation.Detectable](name string) HandlerFunc[T] {
	return func(s *Session[T]) error {
		if s.Response == nil {
			s.Response = NewListTopicResponse(name)
		}
		return nil
	}
}

func SetIDFromPath[T observation.Detectable](allowLatest bool) HandlerFunc[T] {
	return func(s *Session[T]) error {
		idStr := s.Request.PathValue("id")
		if idStr == "" {
			return fmt.Errorf("missing required parameter: id") //this is most likely internal mux setup error
		}

		if allowLatest && idStr == "latest" {
			s.LatestTopic = true
			return nil
		}

		id, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return NewStatusError(fmt.Errorf("invalid id from path"), http.StatusBadRequest)
		}
		s.TopicId = &id
		return nil
	}
}

func Pull[T observation.Detectable](limit int64) HandlerFunc[T] {
	return func(s *Session[T]) error {
		if s.Detector == nil {
			return fmt.Errorf("Topic Query Session missing Detector")
		}

		if s.Response == nil {
			return fmt.Errorf("Topic Query Session missing Response structure")
		}

		//clone directive
		entityDetector := *s.Detector

		//if query uses latest topic. sort and limit to 1
		if s.LatestTopic {
			entityDetector = *entityDetector.SortLatest()
		}

		results, err := entityDetector.Limit(limit).PullAll(s.Request.Context())
		if err != nil {
			return fmt.Errorf("TopicQuery PullAll request error: %v", err)
		}

		//if query uses latest topic. fill the target id in response of the latest topic
		if len(results) > 0 && s.TopicId == nil && s.LatestTopic {
			id := results[0].GetID()
			s.TopicId = &id
			s.Response.SetTargetID(id)
		}

		//if query uses latest topic. fill the target id in response of the latest topic
		if len(results) > 0 && s.TopicId != nil && s.Response.GetTargetID().IsZero() {
			s.Response.SetTargetID(*s.TopicId)
		}

		for _, result := range results {
			s.Response.Append(result)
		}

		return nil
	}
}

func SetEntanglementFrame[T observation.Detectable](frame string) HandlerFunc[T] {
	return func(s *Session[T]) error {
		s.EntangleFrame.SetFrame(frame)
		return nil
	}
}

func CheckEntanglementToken[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		session, err := s.VerifySession()
		if err != nil {
			return NewStatusError(err, http.StatusExpectationFailed)
		}

		if err := s.EntangleFrame.VerifyTokenAlignment(*session); err != nil {
			return NewStatusError(err, http.StatusExpectationFailed)
		}
		return nil
	}
}

func GenerateEntanglement[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		session, err := s.VerifySession()
		if err != nil {
			return NewStatusError(err, http.StatusExpectationFailed)
		}
		s.Response.Append(entanglement.EntangleSession(s.EntangleFrame, *session))
		return nil
	}
}

func CheckBodyCorrelation[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		session, err := s.VerifySession()
		if err != nil {
			return NewStatusError(err, http.StatusExpectationFailed)
		}

		topicBody := any(s.Body)
		entangleTopic, ok := topicBody.(entanglement.Correlatable)
		if !ok {
			log.Printf("Called to CheckBodyCorrelation on uncompatible type %v", s.Body)
			return fmt.Errorf("Error CheckBodyCorrelation")
		}

		return entangleTopic.CheckTransition(entanglement.EntangleSession(s.EntangleFrame, *session))
	}
}
