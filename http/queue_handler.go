package http

import (
	"html/template"
	"net/http"

	"github.com/tanordheim/babyname-tinder"
)

type queueHandler struct {
	nameTemplate      *template.Template
	matchTemplate     *template.Template
	superlikeTemplate *template.Template
	emptyTemplate     *template.Template
	repo              babynames.Repository
}

type matchModel struct {
	Name  string
	Image string
}

type superlikeModel struct {
	Who   string
	Name  string
	Image string
}

type nameModel struct {
	Name               string
	Image              string
	ProgressPercentage int
	DislikedCount      int
}

func newQueueHandler(repo babynames.Repository) *queueHandler {
	return &queueHandler{
		nameTemplate:      parseTemplate("queue_name"),
		matchTemplate:     parseTemplate("queue_match"),
		superlikeTemplate: parseTemplate("queue_superlike"),
		emptyTemplate:     parseTemplate("queue_empty"),
		repo:              repo,
	}
}

func (h *queueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r.Context())

	// If we have a match, show that
	match, err := h.repo.GetAndAcknowledgeUnseenMatch(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if match != "" {
		h.renderMatch(w, r, match)
		return
	}

	// If there's a pending superlike, show that
	like, err := h.repo.GetPendingSuperlike(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if like != "" {
		h.renderSuperlike(w, r, like, babynames.InverseRole(user.Role))
		return
	}

	// If there's a pending name in the queue, show that
	name, dislikes, err := h.repo.GetNextName(r.Context(), user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if name != "" {
		h.renderName(w, r, name, dislikes, user.Role)
		return
	}

	// Show the "out of names" message
	renderTemplate(w, h.emptyTemplate, nil)
}

func (h *queueHandler) renderMatch(w http.ResponseWriter, r *http.Request, name string) {
	model := &matchModel{
		Name:  name,
		Image: getRandomImage(),
	}
	renderTemplate(w, h.matchTemplate, model)
}

func (h *queueHandler) renderSuperlike(w http.ResponseWriter, r *http.Request, name string, role babynames.Role) {
	model := &superlikeModel{
		Who:   babynames.RoleName(role),
		Name:  name,
		Image: getRandomImage(),
	}
	renderTemplate(w, h.superlikeTemplate, model)
}

func (h *queueHandler) renderName(w http.ResponseWriter, r *http.Request, name string, dislikedCount int, role babynames.Role) {
	stats, err := h.repo.GetStats(r.Context(), role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	progressPercentage := int((1.0 - (float64(stats.Queued) / float64(stats.Total))) * 100)
	model := &nameModel{
		Name:               name,
		Image:              getRandomImage(),
		DislikedCount:      dislikedCount,
		ProgressPercentage: progressPercentage,
	}
	renderTemplate(w, h.nameTemplate, model)
}
