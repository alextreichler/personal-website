package handlers

import (
	"net/http"
	"time"

	"github.com/alextreichler/personal-website/internal/auth"
)

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	app.Render(w, r, "login.html", nil)
}

func (app *App) LoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	_, err = auth.Authenticate(app.DB.Conn, username, password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set a simple session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    username, // Storing username for display
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (app *App) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
