package http

import (
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type superlikeHandler struct {
	repo babynames.Repository
}

func newSuperlikeHandler(repo babynames.Repository) *superlikeHandler {
	return &superlikeHandler{
		repo: repo,
	}
}

func (h *superlikeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	name := r.FormValue("name")

	err := h.repo.Superlike(r.Context(), user.Role, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
