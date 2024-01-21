package sitepages

import (
	"errors"
	"fmt"
	"net/http"
)

type WebSessionError struct {
	Message string
	Code    int
}

func (e *WebSessionError) Error() string {
	return fmt.Sprintf("Error 0x%X: %s", e.Code, e.Message)
}

var SESSION_ERROR_NO_SESSION = 1 << 1
var SESSION_ERROR_DECODING_FAILED = 1 << 2
var SESSION_ERROR_SESSION_EXPIRED = 1 << 3
var SESSION_ERROR_IP_MISMATCH = 1 << 4

// VerifySession verifies that the session is valid

func GetAndVerifySession(r *http.Request) (*WebSession, error) {
	// Get the session from the request
	var sessionError *WebSessionError
	session, err := GetRequestSession(r)

	if err != nil {
		errors.As(err, &sessionError)
	}

	if sessionError == nil {
		sessionError = &WebSessionError{}
	}

	if session == nil {
		sessionError.Message += "Session Empty; "
		sessionError.Code |= SESSION_ERROR_NO_SESSION
		return session, sessionError
	}

	// Check if the session is valid

	if session.GetAge() > WEB_SESSION_TTL {
		sessionError.Message += "Session expired; "
		sessionError.Code |= SESSION_ERROR_SESSION_EXPIRED
	}

	realip := GetRealIPFromRequest(r)
	if realip != string(session.FromIp) {
		sessionError.Message += "IP mismatch; "
		sessionError.Code |= SESSION_ERROR_IP_MISMATCH
	}

	if sessionError.Code == 0 {
		return session, nil
	}

	return session, sessionError
}
