package sitepages

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewWebSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("X-Real-IP", "localhost")
	session := newWebSession(GetRealIPFromRequest(r))
	if session.ID == primitive.NilObjectID {
		t.Errorf("NewWebSession: expected session ID to be non-nil")
	}
	if session.GenerateTime.IsZero() {
		t.Errorf("NewWebSession: expected session GenerateTime to be non-zero")
	}

	if !session.GenerateFrom.IsZero() {
		t.Errorf("NewWebSession: expected session GenerateFrom to be zero")
	}
	if session.FirstTime.IsZero() {
		t.Errorf("NewWebSession: expected session FirstTime to be non-zero")
	}

}

func TestRefreshWebSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("X-Real-IP", "localhost")

	session := newWebSession(GetRealIPFromRequest(r))
	newSession := refreshWebSession(GetRealIPFromRequest(r), session)
	if newSession.ID == session.ID {
		t.Errorf("RefreshWebSession: expected new session ID to be different from old session ID")
	}
	if newSession.GenerateTime.IsZero() {
		t.Errorf("RefreshWebSession: expected new session GenerateTime to be non-zero")
	}

	if newSession.GenerateFrom != session.ID {
		t.Errorf("RefreshWebSession: expected new session GenerateFrom to be old session ID")
	}
	if newSession.FirstTime != session.FirstTime {
		t.Errorf("RefreshWebSession: expected new session FirstTime to be old session FirstTime")
	}

}

func TestEncodeSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	session := newWebSession(GetRealIPFromRequest(r))
	encodedSession, err := EncodeSession(*session)
	if err != nil {
		t.Errorf("EncodeSession: expected no error, got %v", err)
	}
	if encodedSession == "" {
		t.Errorf("EncodeSession: expected encoded session to be non-empty")
	}
}

func TestDecodeSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("X-Real-IP", "localhost")

	session := newWebSession(GetRealIPFromRequest(r))
	encodedSession, err := EncodeSession(*session)
	if err != nil {
		t.Errorf("EncodeSession: expected no error, got %v", err)
	}
	decodedSession, err := DecodeSession(encodedSession)
	if err != nil {
		t.Errorf("DecodeSession: expected no error, got %v", err)
	}
	if decodedSession.ID != session.ID {
		t.Errorf("DecodeSession: expected decoded session ID to be equal to original session ID")
	}
	if decodedSession.GenerateTime.Sub(session.GenerateTime).Seconds() > 0 {
		t.Errorf("DecodeSession: expected decoded session GenerateTime to be equal to original session GenerateTime %s %s", decodedSession.GenerateTime.UTC().String(), session.GenerateTime.UTC().String())
	}

	if decodedSession.GenerateFrom != session.GenerateFrom {
		t.Errorf("DecodeSession: expected decoded session GenerateFrom to be equal to original session GenerateFrom")
	}
	if decodedSession.FirstTime.Sub(session.FirstTime).Seconds() > 0 {
		t.Errorf("DecodeSession: expected decoded session FirstTime to be equal to original session FirstTime")
	}

}

// func TestGetRealIPFromRequest(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		request  *http.Request
// 		expected string
// 	}{
// 		{
// 			name:     "no X-Forwarded-For header",
// 			request:  httptest.NewRequest("GET", "/", nil),
// 			expected: "127.0.0.1",
// 		},
// 		{
// 			name:     "X-Forwarded-For header with one IP",
// 			request:  httptest.NewRequest("GET", "/", nil).Header().Set("X-Forwarded-For", "1.2.3.4"),
// 			expected: "1.2.3.4",
// 		},
// 		{
// 			name:     "X-Forwarded-For header with multiple IPs",
// 			request:  httptest.NewRequest("GET", "/", nil).Header().Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8"),
// 			expected: "1.2.3.4",
// 		},
// 		{
// 			name:     "X-Real-IP header",
// 			request:  httptest.NewRequest("GET", "/", nil).Header().Set("X-Real-IP", "9.8.7.6"),
// 			expected: "9.8.7.6",
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			actual := getRealIPFromRequest(test.request)
// 			if actual != test.expected {
// 				t.Errorf("getRealIPFromRequest: expected %s, got %s", test.expected, actual)
// 			}
// 		})
// 	}
// }

// func TestGetHost(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		request  *http.Request
// 		expected string
// 	}{
// 		{
// 			name:     "no Host header",
// 			request:  httptest.NewRequest("GET", "/", nil),
// 			expected: "",
// 		},
// 		{
// 			name:     "Host header with port",
// 			request:  httptest.NewRequest("GET", "/", nil).Header().Set("Host", "example.com:8080"),
// 			expected: "example.com",
// 		},
// 		{
// 			name:     "Host header without port",
// 			request:  httptest.NewRequest("GET", "/", nil).Header().Set("Host", "example.com"),
// 			expected: "example.com",
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			actual := getHost(test.request)
// 			if actual != test.expected {
// 				t.Errorf("getHost: expected %s, got %s", test.expected, actual)
// 			}
// 		})
// 	}
// }

func TestSetNewRequestSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("X-Real-IP", "localhost")

	createdSession := setNewRequestSession(w, GetRealIPFromRequest(r))

	// Check that the cookie was set
	cookies := w.Header().Get("Set-Cookie")
	if cookies == "" {
		t.Errorf("setNewRequestSession: expected cookie to be set")
	}

	// Check that the cookie value is valid
	parts := strings.Split(cookies, ";")
	if len(parts) <= 2 {
		t.Errorf("setNewRequestSession: expected cookie to have more than 2 parts")
	}
	cookieNameValue := strings.Split(parts[0], "=")

	cookieValue := cookieNameValue[1]

	if cookieValue == "" {
		t.Errorf("setNewRequestSession: expected cookie value to be non-empty")
	}

	// Decode the cookie value
	session, err := DecodeSession(cookieValue)
	if err != nil {
		t.Errorf("setNewRequestSession: expected no error when decoding cookie value, got %v", err)
	}

	// Check that the session is valid
	if session.ID == primitive.NilObjectID {
		t.Errorf("setNewRequestSession: expected session ID to be non-nil")
	}
	if session.GenerateTime.IsZero() {
		t.Errorf("setNewRequestSession: expected session GenerateTime to be non-zero")
	}

	if session.GenerateFrom != primitive.NilObjectID {
		t.Errorf("setNewRequestSession: expected session GenerateFrom to be nil")
	}
	if session.FirstTime.IsZero() {
		t.Errorf("setNewRequestSession: expected session FirstTime to be non zero")
	}

	if createdSession.ID != session.ID {
		t.Errorf("setNewRequestSession: expected created session ID to be equal to decoded session ID")
	}

	if createdSession.GenerateTime.Sub(session.GenerateTime).Seconds() > 1 {
		t.Errorf("setNewRequestSession: expected created session GenerateTime to be equal to decoded session GenerateTime diff: %f", session.GenerateTime.Sub(session.GenerateTime).Seconds())
	}

	if createdSession.GenerateFrom != session.GenerateFrom {
		t.Errorf("GenerateFrom not the same")
	}

}
