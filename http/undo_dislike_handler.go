package http

import (
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type undoDislikeHandler struct {
	repo babynames.Repository
}

func newUndoDislikeHandler(repo babynames.Repository) *undoDislikeHandler {
	return &undoDislikeHandler{
		repo: repo,
	}
}

func (h *undoDislikeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	name := r.FormValue("name")

	err := h.repo.UndoDislike(r.Context(), user.Role, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/disliked", http.StatusSeeOther)
}
