package handlers

import (
	"net/http"
)

func (app *App) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	app.Templates.ExecuteTemplate(w, "dashboard.html", nil)
}
