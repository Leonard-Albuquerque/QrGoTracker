package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)
		dur := time.Since(start)
		ip := r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", dur.Milliseconds(),
			"request_id", r.Header.Get("X-Request-Id"),
			"ip_trunc", truncate(ip, 64),
			"ua", truncate(r.UserAgent(), 200),
			"referer", truncate(r.Referer(), 200),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (r *responseWriter) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
