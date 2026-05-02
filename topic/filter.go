package topic

import (
	"fmt"
	"net/http"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/matter/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Filter func(context RequestContext) (*expression.QueryFieldPredicate, error)

func ExpressFilter(context RequestContext, filters ...Filter) ([]expression.QueryFieldPredicate, error) {
	predicates := []expression.QueryFieldPredicate{}

	for _, filter := range filters {
		predicate, err := filter(context)
		if err != nil {
			return nil, err
		}

		if predicate != nil {
			predicates = append(predicates, *predicate)
		}
	}

	return predicates, nil
}

func ByID(allowLatest bool) Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		if !allowLatest && s.TopicId == nil {
			return nil, NewStatusError(fmt.Errorf("invalid id"), http.StatusBadRequest)
		}

		if s.TopicId == nil {
			return nil, nil
		}

		pred := kosmos.Fld("ID").Eq(s.TopicId)
		return &pred, nil
	}
}

func ByRootID(ignoreZero bool) Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		if s.RootId == nil {
			return nil, NewStatusError(fmt.Errorf("invalid id"), http.StatusBadRequest)
		}

		if ignoreZero && s.RootId.IsZero() {
			return nil, nil
		}

		pred := kosmos.Fld("ID").Eq(s.TopicId)
		return &pred, nil
	}
}

type FieldPredicate struct {
	FieldName string
}

func Fld(name string) FieldPredicate {
	return FieldPredicate{name}
}

func (f FieldPredicate) ByIDFromPath(pathName string) Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		idStr := s.Request.PathValue(pathName)
		if idStr == "" {
			return nil, NewStatusString("empty id from path", http.StatusBadRequest)
		}

		id, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, NewStatusString("invalid id from path", http.StatusBadRequest)
		}

		pred := kosmos.Fld(f.FieldName).Eq(id)

		return &pred, nil
	}
}

func (f FieldPredicate) ByIDSetFromQuery(queryName string) Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		values := s.URLQuery()[queryName]

		ids, err := convertStringToIDs(values)
		if err != nil {
			return nil, err
		}
		var pred expression.QueryFieldPredicate
		if len(values) == 1 {
			pred = kosmos.Fld(f.FieldName).Eq(ids[0])
		} else if len(values) > 1 {
			pred = kosmos.Fld(f.FieldName).In(ids)
		}

		return &pred, nil
	}
}

func (f FieldPredicate) ByPathParam(pathName string) Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		value := s.Request.PathValue(pathName)

		pred := kosmos.Fld(f.FieldName).Eq(value)
		return &pred, nil
	}
}

func (f FieldPredicate) ByAuthID(allowUserZero bool) Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		clientSession, err := s.VerifySession()
		if err != nil {
			return nil, NewStatusError(err, http.StatusUnauthorized)
		}

		userid := clientSession.UserId
		if !allowUserZero && userid.IsZero() {
			return nil, NewStatusError(fmt.Errorf("Failed to filter. User id is zero"), http.StatusUnauthorized)
		}

		pred := kosmos.Fld(f.FieldName).Eq(userid)
		return &pred, nil
	}
}

func (f FieldPredicate) ByAuthName() Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		clientSession, err := s.VerifySession()
		if err != nil {
			return nil, NewStatusError(err, http.StatusUnauthorized)
		}

		username := clientSession.UserName
		if clientSession.UserName == "" {
			return nil, NewStatusError(fmt.Errorf("missing required auth parameter: user_name"), http.StatusUnauthorized)
		}

		pred := kosmos.Fld(f.FieldName).Eq(username)

		return &pred, nil
	}
}

func (f FieldPredicate) Eq(value any) Filter {
	return func(s RequestContext) (*expression.QueryFieldPredicate, error) {
		pred := kosmos.Fld(f.FieldName).Eq(value)
		return &pred, nil
	}
}

// type FilterSession struct {
// 	Filters []expression.QueryFieldPredicate
// }

// func (fs *FilterSession) AddFilter(filter ...expression.QueryFieldPredicate) *FilterSession {
// 	fs.Filters = append(fs.Filters, filter...)
// 	return fs
// }

// type FilterFunc func(filter *FilterSession, session *RequestContext) error
// type FilterAccumulator struct {
// 	Pipe []FilterFunc
// }

// func FilterAcc() *FilterAccumulator {
// 	return &FilterAccumulator{}
// }

// func (fa *FilterAccumulator) Chain(chains ...FilterFunc) *FilterAccumulator {
// 	fa.Pipe = append(fa.Pipe, chains...)
// 	return fa
// }

// func (fa *FilterAccumulator) ByID(allowLatest bool) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		if !allowLatest && s.TopicId == nil {
// 			return NewStatusError(fmt.Errorf("invalid id"), http.StatusBadRequest)
// 		}

// 		if s.TopicId == nil {
// 			s.LatestTopic = true
// 			return nil
// 		}

// 		f.AddFilter(kosmos.Fld("ID").Eq(s.TopicId))
// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) ByRootID(ignoreZero bool) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		if s.RootId == nil {
// 			return NewStatusError(fmt.Errorf("invalid id"), http.StatusBadRequest)
// 		}

// 		if ignoreZero && s.RootId.IsZero() {
// 			return nil
// 		}

// 		f.AddFilter(kosmos.Fld("ID").Eq(s.TopicId))
// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) ByString(fieldName string, value string) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		f.AddFilter(
// 			kosmos.Fld(fieldName).Eq(value),
// 		)
// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) ByPathParam(fieldName string, pathName string) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		value := s.Request.PathValue(pathName)

// 		f.AddFilter(
// 			kosmos.Fld(fieldName).Eq(value),
// 		)
// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) ByIDFromPath(fieldName string, pathName string) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		idStr := s.Request.PathValue(pathName)
// 		if idStr == "" {
// 			return fmt.Errorf("empty id from path")
// 		}

// 		id, err := bson.ObjectIDFromHex(idStr)
// 		if err != nil {
// 			return NewStatusError(fmt.Errorf("invalid id from path"), http.StatusBadRequest)
// 		}

// 		f.AddFilter(
// 			kosmos.Fld(fieldName).Eq(id),
// 		)
// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) ByIDSetFromQuery(fieldName string, queryName string) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		values := s.URLQuery()[queryName]

// 		ids, err := convertStringToIDs(values)
// 		if err != nil {
// 			return err
// 		}

// 		if len(values) == 1 {
// 			f.AddFilter(kosmos.Fld(fieldName).Eq(ids[0]))
// 		} else if len(values) > 1 {
// 			f.AddFilter(kosmos.Fld(fieldName).In(ids))
// 		}

// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) AddFilterFromQuery(fieldName string, queryName string) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		values := s.URLQuery()[queryName]

// 		if len(values) == 1 {
// 			f.AddFilter(kosmos.Fld(fieldName).Eq(values[0]))
// 		} else if len(values) > 1 {
// 			f.AddFilter(kosmos.Fld(fieldName).In(values))
// 		}
// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) ByAuthID(fieldName string, allowUserZero bool) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		clientSession, err := s.VerifySession()
// 		if err != nil {
// 			return NewStatusError(err, http.StatusUnauthorized)
// 		}

// 		userid := clientSession.UserId
// 		if !allowUserZero && userid.IsZero() {
// 			return NewStatusError(fmt.Errorf("Failed to filter. User id is zero"), http.StatusUnauthorized)
// 		}

// 		f.AddFilter(kosmos.Fld(fieldName).Eq(userid))
// 		return nil
// 	})
// }

// func (fa *FilterAccumulator) ByAuthName(fieldName string) *FilterAccumulator {
// 	return fa.Chain(func(f *FilterSession, s *RequestContext) error {
// 		clientSession, err := s.VerifySession()
// 		if err != nil {
// 			return NewStatusError(err, http.StatusUnauthorized)
// 		}

// 		username := clientSession.UserName
// 		if clientSession.UserName == "" {
// 			return NewStatusError(fmt.Errorf("missing required auth parameter: user_name"), http.StatusUnauthorized)
// 		}
// 		f.AddFilter(kosmos.Fld(fieldName).Eq(username))

// 		return nil
// 	})
// }

// func (fa FilterAccumulator) Accumulate(s *RequestContext) ([]expression.QueryFieldPredicate, error) {
// 	filter := &FilterSession{}
// 	for _, chainExecution := range fa.Pipe {
// 		if err := chainExecution(filter, s); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return filter.Filters, nil
// }

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
