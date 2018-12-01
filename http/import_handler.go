package http

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/tanordheim/babyname-tinder"
)

type importHandler struct {
	template *template.Template
	repo     babynames.Repository
}

func newImportHandler(repo babynames.Repository) *importHandler {
	return &importHandler{
		template: parseTemplate("import"),
		repo:     repo,
	}
}

func (h *importHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	names := r.FormValue("names")
	names = strings.Replace(names, "\r\n", "\n", -1) // normalize
	nameList := strings.Split(names, "\n")

	importNames := []string{}
	for _, name := range nameList {
		name = strings.TrimSpace(name)
		if name != "" {
			importNames = append(importNames, name)
		}
	}

	err := h.repo.ImportNames(r.Context(), importNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, h.template, len(importNames))
}
