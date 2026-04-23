package topic

import (
	"net/http"
	"testing"

	"github.com/borghives/websession"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestFilterAccumulator_ByID(t *testing.T) {
	objId := bson.NewObjectID()
	tests := []struct {
		name         string
		allowLatest  bool
		topicId      *bson.ObjectID
		expected     bool // true if error is expected
		expectLatest bool
	}{
		{
			name:        "ByID with valid ID",
			allowLatest: false,
			topicId:     &objId,
			expected:    false,
		},
		{
			name:        "ByID without ID, not allowed latest",
			allowLatest: false,
			topicId:     nil,
			expected:    true,
		},
		{
			name:         "ByID without ID, allowed latest",
			allowLatest:  true,
			topicId:      nil,
			expected:     false,
			expectLatest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RequestContext{
				TopicId: tt.topicId,
			}
			f := Filter().ByID(tt.allowLatest)

			filters, err := f.Accumulate(s)
			if tt.expected {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.expectLatest != s.LatestTopic {
					t.Errorf("expected LatestTopic %v, got %v", tt.expectLatest, s.LatestTopic)
				}
				if !tt.expectLatest && len(filters) == 0 {
					t.Errorf("expected filters to be added")
				}
			}
		})
	}
}

func TestFilterAccumulator_ByString(t *testing.T) {
	s := &RequestContext{}
	f := Filter().ByString("test_field", "test_value")

	filters, err := f.Accumulate(s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestFilterAccumulator_ByPathParam(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test/123", nil)
	req.SetPathValue("id", "123")
	s := &RequestContext{
		Request: req,
	}

	f := Filter().ByPathParam("test_field", "id")
	filters, err := f.Accumulate(s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestFilterAccumulator_ByIDFromPath(t *testing.T) {
	req1, _ := http.NewRequest("GET", "/test/invalid", nil)
	req1.SetPathValue("id", "invalid")
	s1 := &RequestContext{Request: req1}
	f1 := Filter().ByIDFromPath("test_field", "id")
	_, err1 := f1.Accumulate(s1)
	if err1 == nil {
		t.Errorf("expected error for invalid hex ID")
	}

	req2, _ := http.NewRequest("GET", "/test/empty", nil)
	req2.SetPathValue("id", "")
	s2 := &RequestContext{Request: req2}
	f2 := Filter().ByIDFromPath("test_field", "id")
	_, err2 := f2.Accumulate(s2)
	if err2 == nil {
		t.Errorf("expected error for empty hex ID")
	}

	validHex := bson.NewObjectID().Hex()
	req3, _ := http.NewRequest("GET", "/test/valid", nil)
	req3.SetPathValue("id", validHex)
	s3 := &RequestContext{Request: req3}
	f3 := Filter().ByIDFromPath("test_field", "id")
	filters, err3 := f3.Accumulate(s3)
	if err3 != nil {
		t.Errorf("unexpected error for valid hex ID: %v", err3)
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestFilterAccumulator_ByIDSetFromQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test?id="+bson.NewObjectID().Hex()+"&id="+bson.NewObjectID().Hex(), nil)
	q := req.URL.Query()
	s := &RequestContext{
		Request:  req,
		urlQuery: &q,
	}

	f := Filter().ByIDSetFromQuery("test_field", "id")
	filters, err := f.Accumulate(s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestFilterAccumulator_AddFilterFromQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test?name=foo&name=bar", nil)
	q := req.URL.Query()
	s := &RequestContext{
		Request:  req,
		urlQuery: &q,
	}

	f := Filter().AddFilterFromQuery("test_field", "name")
	filters, err := f.Accumulate(s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestFilterAccumulator_ByAuthID(t *testing.T) {
	validUserSession := &websession.Session{
		UserId: bson.NewObjectID(),
	}
	s := &RequestContext{
		userSession: validUserSession,
	}

	f := Filter().ByAuthID("test_field", false)
	filters, err := f.Accumulate(s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}

	zeroUserSession := &websession.Session{
		UserId: bson.NilObjectID,
	}
	sZero := &RequestContext{
		userSession: zeroUserSession,
	}
	fZero := Filter().ByAuthID("test_field", false)
	_, errZero := fZero.Accumulate(sZero)
	if errZero == nil {
		t.Errorf("expected error for zero user ID when not allowed")
	}

	fZeroAllowed := Filter().ByAuthID("test_field", true)
	filtersZero, errZeroAllowed := fZeroAllowed.Accumulate(sZero)
	if errZeroAllowed != nil {
		t.Errorf("unexpected error: %v", errZeroAllowed)
	}
	if len(filtersZero) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filtersZero))
	}
}

func TestFilterAccumulator_ByAuthName(t *testing.T) {
	validUserSession := &websession.Session{
		UserName: "testuser",
	}
	s := &RequestContext{
		userSession: validUserSession,
	}

	f := Filter().ByAuthName("test_field")
	filters, err := f.Accumulate(s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}

	emptyUserSession := &websession.Session{
		UserName: "",
	}
	sEmpty := &RequestContext{
		userSession: emptyUserSession,
	}
	fEmpty := Filter().ByAuthName("test_field")
	_, errEmpty := fEmpty.Accumulate(sEmpty)
	if errEmpty == nil {
		t.Errorf("expected error for empty username")
	}
}
