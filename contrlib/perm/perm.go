package perm

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"google.golang.org/appengine"
)

// Session session object
type Session struct {
	ConsumerKey string
	AppID       string
	Domain      string
	State       string
	AccessToken string
}

// LoadSession load session
func LoadSession(r *http.Request) *Session {
	session := &Session{}
	cookie, err := r.Cookie("S")
	if err == nil {
		b, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			_ = json.Unmarshal(b, session)
		}
	}
	return session
}

// SaveSession save session
func SaveSession(w http.ResponseWriter, session *Session) error {
	sessionBytes, err := json.Marshal(session)
	if err != nil {
		return err
	}
	sessionEncoded := base64.StdEncoding.EncodeToString(sessionBytes)

	cookie := &http.Cookie{
		Name:     "S",
		Value:    sessionEncoded,
		Path:     "/",
		Secure:   !appengine.IsDevAppServer(),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	return nil
}
