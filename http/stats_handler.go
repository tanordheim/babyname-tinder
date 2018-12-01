package http

import (
	"html/template"
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type statsHandler struct {
	template *template.Template
	repo     babynames.Repository
}

func newStatsHandler(repo babynames.Repository) *statsHandler {
	return &statsHandler{
		template: parseTemplate("stats"),
		repo:     repo,
	}
}

func (h *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	stats, err := h.repo.GetStats(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, h.template, stats)
}
