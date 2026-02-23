package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		// Allow same-origin for scripts/styles/connect and permit inline for this minimal UI.
		// No external assets are used, so 'self' + 'unsafe-inline' is acceptable here.
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none';")
		next.ServeHTTP(w, r)
	})
}
