package sitepages

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewWebSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	session := NewWebSession(r)
	if session.ID == primitive.NilObjectID {
		t.Errorf("NewWebSession: expected session ID to be non-nil")
	}
	if session.GenerateTime.IsZero() {
		t.Errorf("NewWebSession: expected session GenerateTime to be non-zero")
	}
	if session.GenerateCnt != 1 {
		t.Errorf("NewWebSession: expected session GenerateCnt to be 1")
	}
	if session.GenerateFrom == primitive.NilObjectID {
		t.Errorf("NewWebSession: expected session GenerateFrom to be non-nil")
	}
	if session.FirstTime.IsZero() {
		t.Errorf("NewWebSession: expected session FirstTime to be non-zero")
	}
	if session.FirstIp == "" {
		t.Errorf("NewWebSession: expected session FirstIp to be non-empty")
	}
}

func TestRefreshWebSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	session := NewWebSession(r)
	newSession := RefreshWebSession(session)
	if newSession.ID == session.ID {
		t.Errorf("RefreshWebSession: expected new session ID to be different from old session ID")
	}
	if newSession.GenerateTime.IsZero() {
		t.Errorf("RefreshWebSession: expected new session GenerateTime to be non-zero")
	}
	if newSession.GenerateCnt != session.GenerateCnt+1 {
		t.Errorf("RefreshWebSession: expected new session GenerateCnt to be old session GenerateCnt + 1")
	}
	if newSession.GenerateFrom != session.ID {
		t.Errorf("RefreshWebSession: expected new session GenerateFrom to be old session ID")
	}
	if newSession.FirstTime != session.FirstTime {
		t.Errorf("RefreshWebSession: expected new session FirstTime to be old session FirstTime")
	}
	if newSession.FirstIp != session.FirstIp {
		t.Errorf("RefreshWebSession: expected new session FirstIp to be old session FirstIp")
	}
}

func TestEncodeSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	session := NewWebSession(r)
	encodedSession, err := EncodeSession(session)
	if err != nil {
		t.Errorf("EncodeSession: expected no error, got %v", err)
	}
	if encodedSession == "" {
		t.Errorf("EncodeSession: expected encoded session to be non-empty")
	}
}

func TestDecodeSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	session := NewWebSession(r)
	encodedSession, err := EncodeSession(session)
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
	if decodedSession.GenerateTime != session.GenerateTime {
		t.Errorf("DecodeSession: expected decoded session GenerateTime to be equal to original session GenerateTime")
	}
	if decodedSession.GenerateCnt != session.GenerateCnt {
		t.Errorf("DecodeSession: expected decoded session GenerateCnt to be equal to original session GenerateCnt")
	}
	if decodedSession.GenerateFrom != session.GenerateFrom {
		t.Errorf("DecodeSession: expected decoded session GenerateFrom to be equal to original session GenerateFrom")
	}
	if decodedSession.FirstTime != session.FirstTime {
		t.Errorf("DecodeSession: expected decoded session FirstTime to be equal to original session FirstTime")
	}
	if decodedSession.FirstIp != session.FirstIp {
		t.Errorf("DecodeSession: expected decoded session FirstIp to be equal to original session FirstIp")
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

func TestGetDomain(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "localhost",
			host:     "127.0.0.1",
			expected: "127.0.0.1",
		},
		{
			name:     "example.com",
			host:     "example.com",
			expected: "example.com",
		},
		{
			name:     "subdomain.example.com",
			host:     "subdomain.example.com",
			expected: "example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := getDomain(test.host)
			if actual != test.expected {
				t.Errorf("getDomain: expected %s, got %s", test.expected, actual)
			}
		})
	}
}

func TestSetNewRequestSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	setNewRequestSession(w, r)

	// Check that the cookie was set
	cookies := w.Header().Get("Set-Cookie")
	if cookies == "" {
		t.Errorf("setNewRequestSession: expected cookie to be set")
	}

	// Check that the cookie value is valid
	parts := strings.Split(cookies, ";")
	if len(parts) != 2 {
		t.Errorf("setNewRequestSession: expected cookie to have 2 parts")
	}
	cookieValue := parts[0]
	if cookieValue == "" {
		t.Errorf("setNewRequestSession: expected cookie value to be non-empty")
	}

	// Decode the cookie value
	decodedValue, err := hex.DecodeString(cookieValue)
	if err != nil {
		t.Errorf("setNewRequestSession: expected no error when decoding cookie value, got %v", err)
	}

	// Unmarshal the decoded value into a WebSession
	var session WebSession
	err = bson.Unmarshal(decodedValue, &session)
	if err != nil {
		t.Errorf("setNewRequestSession: expected no error when unmarshaling cookie value, got %v", err)
	}

	// Check that the session is valid
	if session.ID == primitive.NilObjectID {
		t.Errorf("setNewRequestSession: expected session ID to be non-nil")
	}
	if session.GenerateTime.IsZero() {
		t.Errorf("setNewRequestSession: expected session GenerateTime to be non-zero")
	}
	if session.GenerateCnt != 1 {
		t.Errorf("setNewRequestSession: expected session GenerateCnt to be 1")
	}
	if session.GenerateFrom == primitive.NilObjectID {
		t.Errorf("setNewRequestSession: expected session GenerateFrom to be non-nil")
	}
	if session.FirstTime.IsZero() {
		t.Errorf("setNewRequestSession: expected session FirstTime to be non-zero")
	}
	if session.FirstIp == "" {
		t.Errorf("setNewRequestSession: expected session FirstIp to be non-empty")
	}
}

func TestGetRequestSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	setNewRequestSession(w, r)

	// Get the cookie from the request
	cookie, err := r.Cookie("session")
	if err != nil {
		t.Errorf("getRequestSession: expected no error when getting cookie, got %v", err)
	}

	// Decode the cookie value
	decodedValue, err := hex.DecodeString(cookie.Value)
	if err != nil {
		t.Errorf("getRequestSession: expected no error when decoding cookie value, got %v", err)
	}

	// Unmarshal the decoded value into a WebSession
	var session WebSession
	err = bson.Unmarshal(decodedValue, &session)
	if err != nil {
		t.Errorf("getRequestSession: expected no error when unmarshaling cookie value, got %v", err)
	}

	// Check that the session is valid
	if session.ID == primitive.NilObjectID {
		t.Errorf("getRequestSession: expected session ID to be non-nil")
	}
	if session.GenerateTime.IsZero() {
		t.Errorf("getRequestSession: expected session GenerateTime to be non-zero")
	}
	if session.GenerateCnt != 1 {
		t.Errorf("getRequestSession: expected session GenerateCnt to be 1")
	}
	if session.GenerateFrom == primitive.NilObjectID {
		t.Errorf("getRequestSession: expected session GenerateFrom to be non-nil")
	}
	if session.FirstTime.IsZero() {
		t.Errorf("getRequestSession: expected session FirstTime to be non-zero")
	}
	if session.FirstIp == "" {
		t.Errorf("getRequestSession: expected session FirstIp to be non-empty")
	}
}

func TestRefreshRequestSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	setNewRequestSession(w, r)

	// Get the cookie from the request
	cookie, err := r.Cookie("session")
	if err != nil {
		t.Errorf("refreshRequestSession: expected no error when getting cookie, got %v", err)
	}

	// Decode the cookie value
	decodedValue, err := hex.DecodeString(cookie.Value)
	if err != nil {
		t.Errorf("refreshRequestSession: expected no error when decoding cookie value, got %v", err)
	}

	// Unmarshal the decoded value into a WebSession
	var session WebSession
	err = bson.Unmarshal(decodedValue, &session)
	if err != nil {
		t.Errorf("refreshRequestSession: expected no error when unmarshaling cookie value, got %v", err)
	}

	// Refresh the session
	newSession := RefreshRequestSession(w, r)

	// Check that the new session is valid
	if newSession.ID == session.ID {
		t.Errorf("refreshRequestSession: expected new session ID to be different from old session ID")
	}
	if newSession.GenerateTime.IsZero() {
		t.Errorf("refreshRequestSession: expected new session GenerateTime to be non-zero")
	}
	if newSession.GenerateCnt != session.GenerateCnt+1 {
		t.Errorf("refreshRequestSession: expected new session GenerateCnt to be old session GenerateCnt + 1")
	}
	if newSession.GenerateFrom != session.ID {
		t.Errorf("refreshRequestSession: expected new session GenerateFrom to be old session ID")
	}
	if newSession.FirstTime != session.FirstTime {
		t.Errorf("refreshRequestSession: expected new session FirstTime to be old session FirstTime")
	}
	if newSession.FirstIp != session.FirstIp {
		t.Errorf("refreshRequestSession: expected new session FirstIp to be old session FirstIp")
	}
}

func TestGenerateHexIDFromWebSession(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	session := NewWebSession(r)
	hexID := session.GenerateHexID("test")
	fmt.Printf("HexID: %s\n", hexID)
	if hexID == "" {
		t.Errorf("GenerateHexID: expected hex ID to be non-empty")
	}
	if len(hexID) != 24 {
		t.Errorf("GenerateHexID: expected hex ID to be 12 characters long")
	}

	objId, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		t.Errorf("GenerateHexID: expected no error when converting hex ID to ObjectID, got %v", err)
	}
	if objId == primitive.NilObjectID {
		t.Errorf("GenerateHexID: expected ObjectID to be non-nil")
	}

	assert.Equal(t, hexID, objId.Hex())

}
