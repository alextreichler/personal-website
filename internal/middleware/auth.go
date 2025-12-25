package middleware

import (
	"net/http"

	"github.com/alextreichler/personal-website/internal/auth"
)

func AuthMiddleware(isProd bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		_, err = auth.Verify(cookie.Value)
		if err != nil {
			// Invalid signature, clear cookie and redirect
			http.SetCookie(w, &http.Cookie{
				Name:     "admin_session",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   isProd,
				SameSite: http.SameSiteLaxMode,
			})
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
