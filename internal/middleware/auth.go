package middleware

import (
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
