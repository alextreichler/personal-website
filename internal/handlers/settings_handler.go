package handlers

import (
	"log"
	"net/http"
)

func (app *App) AdminEditAbout(w http.ResponseWriter, r *http.Request) {
	aboutContent, err := app.DB.GetSetting("about")
	if err != nil {
		aboutContent = "" // Handle error or empty appropriately
	}

	data := map[string]interface{}{
		"Content": aboutContent,
	}

	app.Render(w, "admin_about.html", data)
}

func (app *App) AdminUpdateAbout(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	err = app.DB.UpdateSetting("about", content)
	if err != nil {
		log.Printf("Error updating about setting: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
