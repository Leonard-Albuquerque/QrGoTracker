package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
	"log/slog"

	"qr-tracker/internal/config"
	"qr-tracker/internal/service"
	"qr-tracker/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LinkHandler struct {
	svc service.LinkService
	cfg *config.Config
	hub *ws.Hub
}

func NewLinkHandler(svc *service.LinkService, cfg *config.Config, hub *ws.Hub) *LinkHandler {
	return &LinkHandler{svc: *svc, cfg: cfg, hub: hub}
}

func (h *LinkHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req struct {
		TargetURL string `json:"target_url"`
		Name      string `json:"name"`
	}
	if err := dec.Decode(&req); err != nil {
		httpErrorJSON(w, http.StatusBadRequest, "invalid_payload", err.Error())
		return
	}

	link, err := h.svc.Create(r.Context(), req.TargetURL, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			httpErrorJSON(w, http.StatusBadRequest, "invalid_url", "target_url precisa ser uma URL http(s) válida")
		case errors.Is(err, service.ErrInvalidCode):
			httpErrorJSON(w, http.StatusBadRequest, "invalid_name", "name deve ter 3-40 caracteres usando letras, números, _ ou -")
		case errors.Is(err, service.ErrCodeUnavailable):
			httpErrorJSON(w, http.StatusConflict, "name_unavailable", "este name já está em uso")
		default:
			httpErrorJSON(w, http.StatusInternalServerError, "error", "server")
		}
		return
	}

	resp := map[string]interface{}{
		"code":        link.Code,
		"short_url":   h.svc.BuildShortURL(link.Code),
		"qr_url":      h.svc.BuildQRURL(link.Code),
		"target_url":  link.TargetURL,
		"click_count": link.Clicks,
		"created_at":  link.CreatedAt.Format(time.RFC3339),
		"is_active":   link.IsActive,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	link, err := h.svc.GetByCode(r.Context(), code)
	if err != nil {
		httpErrorJSON(w, http.StatusInternalServerError, "error", "server")
		return
	}
	if link == nil {
		httpErrorJSON(w, http.StatusNotFound, "not_found", "code not found")
		return
	}
	out := map[string]interface{}{
		"code":        link.Code,
		"target_url":  link.TargetURL,
		"short_url":   h.svc.BuildShortURL(link.Code),
		"qr_url":      h.svc.BuildQRURL(link.Code),
		"click_count": link.Clicks,
		"created_at":  link.CreatedAt.Format(time.RFC3339),
		"is_active":   link.IsActive,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *LinkHandler) GetQR(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	link, err := h.svc.GetByCode(r.Context(), code)
	if err != nil {
		httpErrorJSON(w, http.StatusInternalServerError, "error", "server")
		return
	}
	if link == nil {
		httpErrorJSON(w, http.StatusNotFound, "not_found", "code not found")
		return
	}
	png, err := qrcode.Encode(h.svc.BuildShortURL(code), qrcode.Medium, 256)
	if err != nil {
		httpErrorJSON(w, http.StatusInternalServerError, "error", "qr")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	link, err := h.svc.GetByCode(r.Context(), code)
	if err != nil {
		httpErrorJSON(w, http.StatusInternalServerError, "error", "server")
		return
	}
	if link == nil {
		http.NotFound(w, r)
		return
	}
	if !link.IsActive {
		w.WriteHeader(http.StatusGone)
		_, _ = io.WriteString(w, "gone")
		return
	}
	if err := h.svc.TrackClick(r.Context(), code); err != nil {
		slog.Warn("track click failed", "err", err)
	} else if updated, err := h.svc.GetByCode(r.Context(), code); err == nil && updated != nil {
		h.hub.Broadcast(code, updated.Clicks)
	}
	http.Redirect(w, r, link.TargetURL, http.StatusFound)
}

func (h *LinkHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	link, err := h.svc.GetByCode(r.Context(), code)
	if err != nil {
		httpErrorJSON(w, http.StatusInternalServerError, "error", "server")
		return
	}
	if link == nil {
		httpErrorJSON(w, http.StatusNotFound, "not_found", "code not found")
		return
	}
	out := map[string]interface{}{
		"code":        link.Code,
		"target_url":  link.TargetURL,
		"click_count": link.Clicks,
		"created_at":  link.CreatedAt.Format(time.RFC3339),
		"is_active":   link.IsActive,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *LinkHandler) StatsWS(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hub.Register(code, conn)
	defer h.hub.Unregister(code, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func httpErrorJSON(w http.ResponseWriter, status int, errMsg, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": errMsg, "details": detail})
}
