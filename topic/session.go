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

type RequestContext struct {
	Request        *http.Request
	Response       Response
	EntangleFrame  entanglement.SystemFrame
	TopicId        *bson.ObjectID
	RootId         *bson.ObjectID
	LatestTopic    bool
	urlQuery       *url.Values
	userSession    *websession.Session
	userSessionErr error
}

func NewRequestContext(r *http.Request) *RequestContext {
	return &RequestContext{
		Request: r,
		EntangleFrame: entanglement.Create(
			r.Header.Get("x-entanglement-nonce"),
			r.Header.Get("x-entanglement-token"),
		),
	}
}

func (rs *RequestContext) URLQuery() url.Values {
	if rs.urlQuery == nil {
		q := rs.Request.URL.Query()
		rs.urlQuery = &q
	}
	return *rs.urlQuery
}

func (rs *RequestContext) VerifySession() (*websession.Session, error) {
	if rs.userSession == nil && rs.userSessionErr == nil {
		rs.userSession, rs.userSessionErr = websession.Manager().GetAndVerifySession(rs.Request)
	}

	return rs.userSession, rs.userSessionErr
}

func (rs *RequestContext) GetVerifyEntanglement() (*entanglement.Session, error) {
	session, err := rs.VerifySession()
	if err != nil {
		return nil, NewStatusError(err, http.StatusExpectationFailed)
	}

	if err := rs.EntangleFrame.VerifyTokenAlignment(*session); err != nil {
		return nil, NewStatusError(err, http.StatusExpectationFailed)
	}

	retval := entanglement.EntangleSession(rs.EntangleFrame, *session)
	return &retval, nil

}

type Session[T observation.Detectable] struct {
	RequestContext
	Detector *observation.EntityDetector[T]
	InBody   T
	Output   []any
}

func NewRequestTopicSession[T observation.Detectable](r *http.Request) *Session[T] {
	return &Session[T]{
		RequestContext: *NewRequestContext(r),
	}
}

func (s *Session[T]) TopicDetector() *observation.EntityDetector[T] {
	if s.Detector == nil {
		s.Detector = kosmos.All[T]()
	}

	return s.Detector
}

func (s *Session[T]) DecodeBody() error {
	if s.Request.Body == nil {
		return fmt.Errorf("nil body in request")
	}

	return json.NewDecoder(s.Request.Body).Decode(&s.InBody)
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

func SetRootIDFromPath[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		idStr := s.Request.PathValue("rid")
		if idStr == "" {
			return fmt.Errorf("missing required parameter: rid") //this is most likely internal mux setup error
		}

		id, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return NewStatusError(fmt.Errorf("invalid rid from path"), http.StatusBadRequest)
		}
		s.RootId = &id
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
		}

		//if query uses latest topic. fill the target id in response of the latest topic
		if len(results) > 0 && s.TopicId != nil && s.Response.GetTargetID().IsZero() {
			s.Response.SetTargetID(*s.TopicId)
		}

		for _, result := range results {
			//if root is set and is zero (randomize root)
			if s.RootId != nil && s.RootId.IsZero() {
				renewRoot, ok := any(&result).(Renewable)
				if ok {
					err = renewRoot.Renew()
					if err != nil {
						return NewStatusError(fmt.Errorf("New Page error %v", err), http.StatusBadRequest)
					}
				}
			}

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
		_, err := s.GetVerifyEntanglement()
		return err
	}
}

func GenerateEntanglement[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		session, err := s.GetVerifyEntanglement()
		if err != nil {
			return err
		}
		s.Response.Append(session)
		return nil
	}
}

func CheckInBodyCorrelation[T observation.Detectable]() HandlerFunc[T] {
	return func(s *Session[T]) error {
		session, err := s.GetVerifyEntanglement()
		if err != nil {
			return NewStatusError(err, http.StatusExpectationFailed)
		}

		topicBody := any(s.InBody)
		entangleTopic, ok := topicBody.(entanglement.Correlatable)
		if !ok {
			log.Printf("Called to CheckBodyCorrelation on incompatible type %v", s.InBody)
			return fmt.Errorf("Error CheckBodyCorrelation")
		}

		return entangleTopic.CheckTransition(*session)
	}
}

func CheckAuthenticatedUser[T observation.Detectable](niceExit bool) HandlerFunc[T] {
	return func(s *Session[T]) error {
		session, err := s.VerifySession()
		if err != nil {
			return NewStatusError(err, http.StatusExpectationFailed)
		}

		if session.UserId.IsZero() || session.UserName == "" {
			if niceExit {
				return NewStatusError(fmt.Errorf("No user early nice exit"), http.StatusAccepted)
			}
			return NewStatusError(fmt.Errorf("No User"), http.StatusUnauthorized)
		}

		return nil
	}
}
