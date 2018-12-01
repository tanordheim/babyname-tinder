package http

import (
	"encoding/csv"
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type exportMatchesHandler struct {
	repo babynames.Repository
}

func newExportMatchesHandler(repo babynames.Repository) *exportMatchesHandler {
	return &exportMatchesHandler{
		repo: repo,
	}
}

func (h *exportMatchesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	matches, err := h.repo.GetMatches(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/csv")
	w.Header().Add("Content-Disposition", "attachment; filename=\"matches.csv\"")

	csv := csv.NewWriter(w)
	defer csv.Flush()

	csv.Write([]string{
		"Name",
		"Dad Superliked",
		"Mom Superliked",
	})

	for _, match := range matches {
		dadSuperliked := "0"
		momSuperliked := "0"
		for rn, r := range match.Roles {
			if r.Superliked && rn == babynames.DadRole {
				dadSuperliked = "1"
			}
			if r.Superliked && rn == babynames.MomRole {
				momSuperliked = "1"
			}
		}

		csv.Write([]string{
			match.Name,
			dadSuperliked,
			momSuperliked,
		})
	}
}
