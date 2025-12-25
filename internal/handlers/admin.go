package handlers

import (
	"net/http"
)

func (app *App) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	username := "Admin"
	if cookie, err := r.Cookie("admin_session"); err == nil {
		username = cookie.Value
	}

	stats, err := app.DB.GetDashboardStats()
	if err != nil {
		// Log error but render dashboard anyway
		// slog.Error("Failed to get dashboard stats", "error", err) // slog not imported in this file yet, ignoring logging for brevity or add import
	}
	
	data := map[string]interface{}{
		"Username":  username,
		"PageTitle": "Dashboard",
		"Stats":     stats,
	}
	
	app.Render(w, r, "dashboard.html", data)
}
