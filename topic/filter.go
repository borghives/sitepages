package topic

import (
	"fmt"
	"net/http"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
)


type FilterSession struct {
	Filters []expression.QueryFieldPredicate
}

func (fs *FilterSession) AddFilter(filter ...expression.QueryFieldPredicate) *FilterSession {
	fs.Filters = append(fs.Filters, filter...)
	return fs
}

type FilterFunc func(filter *FilterSession, session *RequestSession) error
type FilterAccumulator struct {
	Pipe []FilterFunc
}

func Filter() *FilterAccumulator {
	return &FilterAccumulator{}
}

func (fa *FilterAccumulator) Chain(chains ...FilterFunc) *FilterAccumulator {
	fa.Pipe = append(fa.Pipe, chains...)
	return fa
}

func (fa *FilterAccumulator) ByID(allowLatest bool) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		if !allowLatest && s.TopicId == nil {
			return NewStatusError(fmt.Errorf("invalid id"), http.StatusBadRequest)
		}

		if s.TopicId == nil {
			s.LatestTopic = true
			return nil
		}

		f.AddFilter(kosmos.Fld("ID").Eq(s.TopicId))
		return nil
	})
}

func (fa *FilterAccumulator) ByString(fieldName string, value string) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		f.AddFilter(
			kosmos.Fld(fieldName).Eq(value),
		)
		return nil
	})
}

func (fa *FilterAccumulator) ByPathParam(fieldName string, pathName string) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		value := s.Request.PathValue(pathName)

		f.AddFilter(
			kosmos.Fld(fieldName).Eq(value),
		)
		return nil
	})
}

func (fa *FilterAccumulator) ByIDFromPath(fieldName string, pathName string) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		idStr := s.Request.PathValue(pathName)
		if idStr == "" {
			return fmt.Errorf("empty id from path")
		}

		id, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return NewStatusError(fmt.Errorf("invalid id from path"), http.StatusBadRequest)
		}

		f.AddFilter(
			kosmos.Fld(fieldName).Eq(id),
		)
		return nil
	})
}

func (fa *FilterAccumulator) ByIDSetFromQuery(fieldName string, queryName string) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		values := s.URLQuery()[queryName]

		ids, err := convertStringToIDs(values)
		if err != nil {
			return err
		}

		if len(values) == 1 {
			f.AddFilter(kosmos.Fld(fieldName).Eq(ids[0]))
		} else if len(values) > 1 {
			f.AddFilter(kosmos.Fld(fieldName).In(ids))
		}

		return nil
	})
}

func (fa *FilterAccumulator) AddFilterFromQuery(fieldName string, queryName string) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		values := s.URLQuery()[queryName]

		if len(values) == 1 {
			f.AddFilter(kosmos.Fld(fieldName).Eq(values[0]))
		} else if len(values) > 1 {
			f.AddFilter(kosmos.Fld(fieldName).In(values))
		}
		return nil
	})
}

func (fa *FilterAccumulator) ByAuthID(fieldName string, allowUserZero bool) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		clientSession, err := s.VerifySession()
		if err != nil {
			return NewStatusError(err, http.StatusUnauthorized)
		}

		userid := clientSession.UserId
		if !allowUserZero && userid.IsZero() {
			return NewStatusError(fmt.Errorf("Failed to filter. User id is zero"), http.StatusUnauthorized)
		}

		f.AddFilter(kosmos.Fld(fieldName).Eq(userid))
		return nil
	})
}

func (fa *FilterAccumulator) ByAuthName(fieldName string) *FilterAccumulator {
	return fa.Chain(func(f *FilterSession, s *RequestSession) error {
		clientSession, err := s.VerifySession()
		if err != nil {
			return NewStatusError(err, http.StatusUnauthorized)
		}

		username := clientSession.UserName
		if clientSession.UserName == "" {
			return NewStatusError(fmt.Errorf("missing required auth parameter: user_name"), http.StatusUnauthorized)
		}
		f.AddFilter(kosmos.Fld(fieldName).Eq(username))

		return nil
	})
}

func (fa FilterAccumulator) Accumulate(s *RequestSession) ([]expression.QueryFieldPredicate, error) {
	filter := &FilterSession{}
	for _, chainExecution := range fa.Pipe {
		if err := chainExecution(filter, s); err != nil {
			return nil, err
		}
	}
	return filter.Filters, nil
}

func convertStringToIDs(idStrs []string) ([]bson.ObjectID, error) {
	ids := make([]bson.ObjectID, 0)
	for _, idStr := range idStrs {
		id, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
