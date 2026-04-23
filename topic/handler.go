package topic

import (
	"log"
	"net/http"

	"github.com/borghives/kosmos-go/observation"
)

type HandlerFunc[T observation.Detectable] func(session *Session[T]) error
type Handler[T observation.Detectable] struct {
	Pipe []HandlerFunc[T]
}

func (h *Handler[T]) Chain(chains ...HandlerFunc[T]) *Handler[T] {
	h.Pipe = append(h.Pipe, chains...)
	return h
}

func (h Handler[T]) AggregateSession(r *http.Request) (*Session[T], error) {
	session := NewRequestTopicSession[T](r)
	for _, chainExecution := range h.Pipe {
		if err := chainExecution(session); err != nil {
			return nil, err
		}
	}
	return session, nil
}

var MAX_BODY_SIZE int64 = 1048576

func (t *Handler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, MAX_BODY_SIZE)
	}

	session, err := t.AggregateSession(r)
	if err != nil {
		ServeError(w, err)
		return
	}
	MarshalResponse(session.Response, w)
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
