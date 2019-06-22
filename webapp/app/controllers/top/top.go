package top

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo"
	"github.com/s-aska/line-works-debugger-go/contrlib/perm"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// Root top page
func Root(c echo.Context) error {
	r := c.Request()

	session := perm.LoadSession(r)

	data := map[string]interface{}{}
	data["req"] = r
	data["session"] = session
	return c.Render(http.StatusOK, "index.html", data)
}

// Post generate token
func Post(c echo.Context) error {
	r := c.Request()
	w := c.Response().Writer

	b, err := generateRandomBytes(32)
	if err != nil {
		return err
	}
	state := hex.EncodeToString(b)

	session := &perm.Session{
		ConsumerKey: r.FormValue("ConsumerKey"),
		AppID:       r.FormValue("AppID"),
		Domain:      r.FormValue("Domain"),
		State:       state,
	}
	err = perm.SaveSession(w, session)
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("client_id", session.ConsumerKey)
	q.Set("redirect_uri", scheme()+"://"+r.Host+r.RequestURI+"callback")
	q.Set("state", session.State)
	q.Set("domain", session.Domain)
	redirectURL := fmt.Sprintf("https://auth.worksmobile.com/ba/%s/service/authorize?%s", session.AppID, q.Encode())
	return c.Redirect(http.StatusFound, redirectURL)
}

// Callback callback page
func Callback(c echo.Context) error {
	r := c.Request()
	w := c.Response().Writer
	ctx := r.Context()

	session := perm.LoadSession(r)

	code := r.FormValue("code")
	state := r.FormValue("state")
	if session.State != state {
		return errors.New("invalid state")
	}

	values := url.Values{}
	values.Add("client_id", session.ConsumerKey)
	values.Add("domain", session.Domain)
	values.Add("code", code)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://auth.worksmobile.com/ba/%s/service/token", session.AppID), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type token struct {
		ErrorCode    string `json:"errorCode"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpireIn     string `json:"expire_in"`
	}
	t := &token{}
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		return err
	}
	if t.AccessToken == "" {
		return fmt.Errorf("Error: %+v", t)
	}

	session.AccessToken = t.AccessToken
	err = perm.SaveSession(w, session)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func scheme() string {
	if appengine.IsDevAppServer() {
		return "http"
	}
	return "https"
}
