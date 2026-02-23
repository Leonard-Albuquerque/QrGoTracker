package middleware

import (
	"net/http"
	"sync"
	"time"

	"qr-tracker/internal/config"
)

type limiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rps     int
	burst   int
}

type tokenBucket struct {
	tokens int
	last   time.Time
}

func newLimiter(rps, burst int) *limiter {
	return &limiter{buckets: make(map[string]*tokenBucket), rps: rps, burst: burst}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	tb, ok := l.buckets[key]
	now := time.Now()
	if !ok {
		l.buckets[key] = &tokenBucket{tokens: l.burst, last: now}
		tb = l.buckets[key]
	}
	elapsed := now.Sub(tb.last).Seconds()
	add := int(elapsed * float64(l.rps))
	if add > 0 {
		tb.tokens += add
		if tb.tokens > l.burst {
			tb.tokens = l.burst
		}
		tb.last = now
	}
	if tb.tokens > 0 {
		tb.tokens -= 1
		return true
	}
	return false
}

func RateLimit(cfg *config.Config) func(http.Handler) http.Handler {
	l := newLimiter(cfg.RL_RPS, cfg.RL_BURST)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.RL_Enabled {
				next.ServeHTTP(w, r)
				return
			}
			ip := r.RemoteAddr
			if !l.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
