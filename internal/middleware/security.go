package middleware

import (
	"net/http"
)

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Protects against MIME sniffing vulnerabilities
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevents the site from being embedded in an iframe (clickjacking protection)
		w.Header().Set("X-Frame-Options", "DENY")
		// Controls how much referrer information is sent
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Basic XSS protection (for older browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		next.ServeHTTP(w, r)
	})
}
