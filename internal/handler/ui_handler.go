package handler

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"qr-tracker/internal/config"
)

type UIHandler struct {
	tmpl *template.Template
	cfg  *config.Config
}

func NewUIHandler(t *template.Template, cfg *config.Config) *UIHandler {
	return &UIHandler{tmpl: t, cfg: cfg}
}

func (u *UIHandler) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = u.tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{"BaseURL": u.cfg.BaseURL})
}

func (u *UIHandler) Stats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = u.tmpl.ExecuteTemplate(w, "stats.html", map[string]interface{}{"BaseURL": u.cfg.BaseURL, "Code": code})
}
