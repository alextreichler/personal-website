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
		
		// Re-render the login page with an error message
		data := map[string]interface{}{
			"PageTitle": "Login",
			"Error":     "Invalid credentials",
		}
		app.Render(w, r, "login.html", data)
		return
	}

	// Set a signed session cookie
	signedValue := auth.Sign(username)
	http.SetCookie(w, &http.Cookie{
		Name:     app.Config.SessionCookie,
		Value:    signedValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.Config.Env == "production",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (app *App) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     app.Config.SessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.Config.Env == "production",
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
