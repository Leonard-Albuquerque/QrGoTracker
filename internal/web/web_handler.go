package web

import (
    "embed"
    "html/template"
    "net/http"

    "github.com/go-chi/chi/v5"

    "qr-tracker/internal/config"
)

//go:embed templates/*.html
var templatesFS embed.FS

type WebHandler struct {
    tmpl *template.Template
    cfg  *config.Config
}

func NewWebHandler(cfg *config.Config) *WebHandler {
    t := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
    return &WebHandler{tmpl: t, cfg: cfg}
}

func (w *WebHandler) IndexPage(rw http.ResponseWriter, r *http.Request) {
    rw.Header().Set("Content-Type", "text/html; charset=utf-8")
    _ = w.tmpl.ExecuteTemplate(rw, "index.html", map[string]interface{}{"BaseURL": w.cfg.BaseURL})
}

func (w *WebHandler) StatsPage(rw http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    rw.Header().Set("Content-Type", "text/html; charset=utf-8")
    _ = w.tmpl.ExecuteTemplate(rw, "stats.html", map[string]interface{}{"BaseURL": w.cfg.BaseURL, "Code": code})
}
