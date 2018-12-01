package http

import (
	"html/template"
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type likedHandler struct {
	template *template.Template
	repo     babynames.Repository
}

func newLikedHandler(repo babynames.Repository) *likedHandler {
	return &likedHandler{
		template: parseTemplate("liked"),
		repo:     repo,
	}
}

func (h *likedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	likes, err := h.repo.GetLikedNames(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, h.template, likes)
}
