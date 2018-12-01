package http

import (
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type likeHandler struct {
	repo babynames.Repository
}

func newLikeHandler(repo babynames.Repository) *likeHandler {
	return &likeHandler{
		repo: repo,
	}
}

func (h *likeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	name := r.FormValue("name")

	err := h.repo.Like(r.Context(), user.Role, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
