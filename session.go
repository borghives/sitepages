package sitepages

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var WEB_SESSION_TTL = time.Hour * 12

type WebSession struct {
	ID           primitive.ObjectID `json:"ID" bson:"_id,omitempty"`
	FromIp       string             `json:"ip" bson:"ip"`
	GenerateTime time.Time          `json:"GenTm" bson:"gen_tm"`
	GenerateFrom primitive.ObjectID `json:"GenFrm" bson:"gen_frm"`
	FirstID      primitive.ObjectID `json:"FrstID" bson:"frst_id"`
	FirstTime    time.Time          `json:"FrstTm" bson:"frst_tm"`
}

func NewWebSession(realIP string) *WebSession {
	currentTime := time.Now()
	id := primitive.NewObjectID()

	return &WebSession{
		ID:           id,
		FromIp:       realIP,
		GenerateTime: currentTime,
		FirstID:      id,
		FirstTime:    currentTime,
	}
}

func RefreshWebSession(realIP string, oldSession *WebSession) *WebSession {
	return &WebSession{
		ID:           primitive.NewObjectID(),
		FromIp:       realIP,
		GenerateTime: time.Now(),
		GenerateFrom: oldSession.ID,
		FirstID:      oldSession.FirstID,
		FirstTime:    oldSession.FirstTime,
	}
}

func EncodeSession(session WebSession) (string, error) {
	secret := os.Getenv("SESSION_LATEST")

	encodedBytes, err := bson.Marshal(session)
	if err != nil {
		return "", err
	}
	return EncryptMessage(secret, encodedBytes)
}

func DecodeSession(encodedSession string) (*WebSession, error) {
	secret := os.Getenv("SESSION_LATEST")
	decodedBytes, err := DecryptMessage(secret, encodedSession)
	if err != nil {
		return nil, err
	}
	var session WebSession
	err = bson.Unmarshal(decodedBytes, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// getRealIPFromRequest extracts the client's real IP address from http.Request
func GetRealIPFromRequest(r *http.Request) string {
	// Check the X-Forwarded-For header first
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// This header can contain multiple IPs separated by comma
		// The first one in the list is the original client IP
		parts := strings.Split(xForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts[0]
	}

	// If X-Forwarded-For is empty, check the X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// If neither header is present, use the remote address from the request
	// This might be the IP of a proxy or load balancer
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func getDomain() string {
	domain := os.Getenv("SITE_DOMAIN")
	if domain == "" {
		domain = "127.0.0.1"
	}
	return domain
}

func setNewRequestSession(w http.ResponseWriter, realIP string) *WebSession {

	// Create a new session
	session := NewWebSession(realIP)
	setSessionCookie(w, session)
	return session
}

func setSessionCookie(w http.ResponseWriter, session *WebSession) error {
	domain := getDomain()

	// Create a new session
	encodedSess, err := EncodeSession(*session)
	if err != nil {
		return err
	}
	// Create a new cookie
	cookie := http.Cookie{
		Name:     "session",
		Value:    encodedSess,
		Path:     "/",     // The cookie is accessible on all paths
		Domain:   domain,  // Accessible by mypierian.com and all its subdomains
		MaxAge:   1469000, // Expires after ~17 days
		HttpOnly: true,    // Not accessible via JavaScript
	}

	// Set the cookie in the response header
	http.SetCookie(w, &cookie)
	return nil
}

func GetRequestSession(r *http.Request) (*WebSession, error) {
	// Get the cookie from the request
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, &WebSessionError{
			Message: "No session found",
			Code:    SESSION_ERROR_NO_SESSION,
		}
	}

	// Decode the cookie value
	session, err := DecodeSession(cookie.Value)
	if err != nil {
		return nil, &WebSessionError{
			Message: "FAILED to decode session",
			Code:    SESSION_ERROR_DECODING_FAILED,
		}
	}

	// Return the decoded session
	return session, nil
}

func RefreshRequestSession(w http.ResponseWriter, r *http.Request) *WebSession {
	// Get the session from the request
	session, _ := GetRequestSession(r)
	if session == nil {
		return setNewRequestSession(w, GetRealIPFromRequest(r))
	}

	return session
}

func (sess WebSession) GetAge() time.Duration {
	return time.Since(sess.GenerateTime)
}

func (sess *WebSession) GenerateHexID(message string) string {
	if sess == nil {
		return primitive.ObjectID{}.Hex()
	}
	idbytes := sha256.Sum256([]byte(sess.ID.Hex() + message))
	//convert bytes to hex string
	return string(hex.EncodeToString(idbytes[:12]))
}
