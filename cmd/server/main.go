package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"qr-tracker/internal/config"
	"qr-tracker/internal/database"
	"qr-tracker/internal/handler"
	"qr-tracker/internal/middleware"
	"qr-tracker/internal/repository"
	"qr-tracker/internal/service"

	"qr-tracker/internal/web"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.LoadFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	db := database.MustConnect(cfg)
	repo := repository.NewSQLiteLinkRepository(db)
	svc := service.NewLinkService(repo, cfg)
	h := handler.NewLinkHandler(svc, cfg)
	// init web UI handler (server-side templates embedded)
	webHandler := web.NewWebHandler(cfg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.RateLimit(cfg))

	r.Get("/healthz", h.Health)
	r.Get("/", webHandler.IndexPage)
	r.Post("/api/links", h.CreateLink)
	r.Get("/api/links/{code}/stats", h.GetStats)
	r.Get("/qr/{code}.png", h.GetQR)
	r.Get("/r/{code}", h.Redirect)
	r.Get("/stats/{code}", webHandler.StatsPage)
	r.Handle("/assets/*", webHandler.AssetsHandler())

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	slog.Info("server stopped")
}
