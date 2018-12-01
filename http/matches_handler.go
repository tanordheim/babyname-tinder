package http

import (
	"html/template"
	"net/http"
	"time"

	"github.com/tanordheim/babyname-tinder"
)

type matchesHandler struct {
	template *template.Template
	repo     babynames.Repository
}

type matchesModel struct {
	Name          string
	MatchedAt     time.Time
	MomSuperliked bool
	DadSuperliked bool
}

func newMatchesHandler(repo babynames.Repository) *matchesHandler {
	return &matchesHandler{
		template: parseTemplate("matches"),
		repo:     repo,
	}
}

func (h *matchesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())
	matches, err := h.repo.GetMatches(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]*matchesModel, len(matches))
	for idx, match := range matches {
		matchedAt := time.Time{}
		dadSuperliked := false
		momSuperliked := false
		for rn, r := range match.Roles {
			if r.LikedAt.After(matchedAt) {
				matchedAt = r.LikedAt
			}
			if r.Superliked && rn == babynames.DadRole {
				dadSuperliked = true
			}
			if r.Superliked && rn == babynames.MomRole {
				momSuperliked = true
			}
		}
		res[idx] = &matchesModel{
			Name:          match.Name,
			MatchedAt:     matchedAt,
			DadSuperliked: dadSuperliked,
			MomSuperliked: momSuperliked,
		}
	}
	renderTemplate(w, h.template, res)
}
