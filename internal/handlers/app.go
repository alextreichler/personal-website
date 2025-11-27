package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/alextreichler/personal-website/internal/repository"
)

type App struct {
	DB        *repository.Database
	Templates *template.Template
}

func NewApp(db *repository.Database) *App {
	// Parse templates
	// Using Must instead of ParseGlob to catch errors at startup.
	// This approach loads all templates at once, suitable for smaller sites.
	tmpl := template.Must(template.ParseFiles(
		"web/template/base.html",
		"web/template/home.html",
		"web/template/login.html",
		// Add other templates here as they are created
	))


	return &App{
		DB:        db,
		Templates: tmpl,
	}
}

func (app *App) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"CurrentYear": time.Now().Year(),
	}

	err := app.Templates.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}