package web

import (
	"embed"
	"html/template"
	iofs "io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	"qr-tracker/internal/config"
)

//go:embed templates/*.html templates/assets/*
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

// AssetsHandler serves embedded static assets from templates/assets
func (w *WebHandler) AssetsHandler() http.Handler {
	sub, err := iofs.Sub(templatesFS, "templates/assets")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.StripPrefix("/assets/", http.FileServer(http.FS(sub)))
}
