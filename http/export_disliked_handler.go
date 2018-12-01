package http

import (
	"encoding/csv"
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type exportDislikedHandler struct {
	repo babynames.Repository
}

func newExportDislikedHandler(repo babynames.Repository) *exportDislikedHandler {
	return &exportDislikedHandler{
		repo: repo,
	}
}

func (h *exportDislikedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	dislikes, err := h.repo.GetDislikedNames(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/csv")
	w.Header().Add("Content-Disposition", "attachment; filename=\"dislikes.csv\"")

	csv := csv.NewWriter(w)
	defer csv.Flush()

	csv.Write([]string{"Name"})

	for _, match := range dislikes {
		csv.Write([]string{match.Name})
	}
}
