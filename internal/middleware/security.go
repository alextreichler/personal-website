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

		// Content Security Policy (CSP)
		// Allows necessary CDNs for Fonts, Icons (FontAwesome), and Editor (EasyMDE)
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com https://cdn.jsdelivr.net; img-src 'self' data:; font-src 'self' https://fonts.gstatic.com https://cdnjs.cloudflare.com; connect-src 'self';")

		// Strict Transport Security (HSTS)
		// Tells browsers to cache the fact that this site should only be accessed via HTTPS for the next 2 years
		// Note: Only effective if the site is actually served over HTTPS
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
