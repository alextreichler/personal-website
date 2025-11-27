package handlers

import (
	"net/http"
)

func (app *App) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	username := "Admin"
	if cookie, err := r.Cookie("admin_session"); err == nil {
		username = cookie.Value
	}
	
	data := map[string]interface{}{
		"Username": username,
	}
	
	app.Render(w, r, "dashboard.html", data)
}
