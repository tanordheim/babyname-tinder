package http

import (
	"encoding/csv"
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type exportLikedHandler struct {
	repo babynames.Repository
}

func newExportLikedHandler(repo babynames.Repository) *exportLikedHandler {
	return &exportLikedHandler{
		repo: repo,
	}
}

func (h *exportLikedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	likes, err := h.repo.GetLikedNames(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/csv")
	w.Header().Add("Content-Disposition", "attachment; filename=\"likes.csv\"")

	csv := csv.NewWriter(w)
	defer csv.Flush()

	csv.Write([]string{"Name"})

	for _, match := range likes {
		csv.Write([]string{match.Name})
	}
}
