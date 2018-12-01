package http

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type loginHandler struct {
	config  *oauth2.Config
	session sessions.Store
}

func newLoginHandler(config *oauth2.Config, session sessions.Store) *loginHandler {
	return &loginHandler{
		config:  config,
		session: session,
	}
}

func (h *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Generate some random state and put it in the session
	buf := make([]byte, 32)
	rand.Read(buf)
	state := base64.StdEncoding.EncodeToString(buf)

	session, err := h.session.Get(r, "state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start the auth flow
	aud := oauth2.SetAuthURLParam("audience", fmt.Sprintf("https://%s/userinfo", auth0Domain))
	url := h.config.AuthCodeURL(state, aud)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
