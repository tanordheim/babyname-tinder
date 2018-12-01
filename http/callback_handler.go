package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type callbackHandler struct {
	config  *oauth2.Config
	session sessions.Store
}

func newCallbackHandler(config *oauth2.Config, session sessions.Store) *callbackHandler {
	return &callbackHandler{
		config:  config,
		session: session,
	}
}

func (h *callbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	session, err := h.session.Get(r, "state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if state != session.Values["state"] {
		http.Error(w, "Invalid state parameter", http.StatusInternalServerError)
		return
	}

	// Swap auth code for token
	code := r.URL.Query().Get("code")
	token, err := h.config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Figure out who the user is
	client := h.config.Client(r.Context(), token)
	resp, err := client.Get(fmt.Sprintf("https://%s/userinfo", auth0Domain))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var profile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store user info in the session
	session, err = h.session.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	emailAddress := profile["email"].(string)
	role, err := getRoleForEmail(emailAddress)
	if err != nil {
		http.Error(w, "Unauthorized user", http.StatusUnauthorized)
		return
	}
	session.Values["user"] = newUser(emailAddress, role)

	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the front page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
