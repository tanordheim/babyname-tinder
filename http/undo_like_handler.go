package http

import (
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type undoLikeHandler struct {
	repo babynames.Repository
}

func newUndoLikeHandler(repo babynames.Repository) *undoLikeHandler {
	return &undoLikeHandler{
		repo: repo,
	}
}

func (h *undoLikeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	name := r.FormValue("name")

	err := h.repo.UndoLike(r.Context(), user.Role, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/liked", http.StatusSeeOther)
}
