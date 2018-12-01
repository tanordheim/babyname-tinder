package http

import (
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type dislikeHandler struct {
	repo babynames.Repository
}

func newDislikeHandler(repo babynames.Repository) *dislikeHandler {
	return &dislikeHandler{
		repo: repo,
	}
}

func (h *dislikeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	name := r.FormValue("name")

	_, err := h.repo.Dislike(r.Context(), user.Role, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
