package middleware

    import (
        "context"
        "crypto/rand"
        "encoding/hex"
        "net/http"
    )

    type ctxKeyRequestID struct{}

    func getRandomID() string {
        b := make([]byte, 8)
        _, _ = rand.Read(b)
        return hex.EncodeToString(b)
    }

    func RequestID(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            id := r.Header.Get("X-Request-Id")
            if id == "" {
                id = getRandomID()
            }
            ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, id)
            w.Header().Set("X-Request-Id", id)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
