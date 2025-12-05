package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

type key int

const (
	csrfTokenKey key = iota
)

// CSRFMiddleware handles CSRF protection by ensuring a valid token is present
// in state-changing requests and making the token available to handlers.
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get or Create Token from Cookie
		token := ""
		cookie, err := r.Cookie("csrf_token")
		if err == nil {
			token = cookie.Value
		}

		if token == "" {
			randomBytes := make([]byte, 32)
			rand.Read(randomBytes)
			token = base64.URLEncoding.EncodeToString(randomBytes)

			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		}

		// 2. If State Changing, Verify Token
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH" {
			// ParseMultipartForm might be necessary for upload forms,
			// but FormValue usually handles it if the content-type is set correctly.
			// We'll try to retrieve it from FormValue or Header.
			sentToken := r.FormValue("csrf_token")
			if sentToken == "" {
				sentToken = r.Header.Get("X-CSRF-Token")
			}

			if sentToken == "" || sentToken != token {
				http.Error(w, "Forbidden - CSRF token mismatch", http.StatusForbidden)
				return
			}
		}

		// 3. Add Token to Context for Handlers/Templates
		ctx := context.WithValue(r.Context(), csrfTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetCSRFToken(r *http.Request) string {
	if val, ok := r.Context().Value(csrfTokenKey).(string); ok {
		return val
	}
	return ""
}
