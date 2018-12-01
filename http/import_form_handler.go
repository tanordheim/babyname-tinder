package http

import (
	"html/template"
	"net/http"
)

type importFormHandler struct {
	template *template.Template
}

func newImportFormHandler() *importFormHandler {
	return &importFormHandler{
		template: parseTemplate("import_form"),
	}
}

func (h *importFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, h.template, nil)
}
