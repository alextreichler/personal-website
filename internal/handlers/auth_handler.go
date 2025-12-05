package handlers

import (
	"net/http"
	"time"

	"github.com/alextreichler/personal-website/internal/auth"
)

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"PageTitle": "Login",
	}
	app.Render(w, r, "login.html", data)
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
		time.Sleep(2 * time.Second) // Mitigate brute-force attacks
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set a signed session cookie
	signedValue := auth.Sign(username)
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    signedValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Added for security
		SameSite: http.SameSiteLaxMode, // Added for CSRF protection
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
		Secure:   true, // Added for security
		SameSite: http.SameSiteLaxMode, // Added for CSRF protection
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
