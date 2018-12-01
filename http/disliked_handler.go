package http

import (
	"html/template"
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type dislikedHandler struct {
	template *template.Template
	repo     babynames.Repository
}

func newDislikedHandler(repo babynames.Repository) *dislikedHandler {
	return &dislikedHandler{
		template: parseTemplate("disliked"),
		repo:     repo,
	}
}

func (h *dislikedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	dislikes, err := h.repo.GetDislikedNames(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, h.template, dislikes)
}
