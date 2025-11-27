package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alextreichler/personal-website/internal/repository"
	"github.com/yuin/goldmark"
)

type App struct {
	DB            *repository.Database
	TemplateCache map[string]*template.Template
}

func NewApp(db *repository.Database) *App {
	// Pre-compile templates into a cache
	cache := make(map[string]*template.Template)

	pages := []string{
		"home.html",
		"login.html",
		"dashboard.html",
		"admin_posts.html",
		"admin_post_new.html",
		"admin_post_edit.html",
		"admin_about.html",
		"post.html",
	}

	for _, page := range pages {
		name := page
		files := []string{
			"web/template/base.html",
			filepath.Join("web/template", page),
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			log.Fatalf("Error parsing template %s: %v", name, err)
		}

		cache[name] = ts
	}

	return &App{
		DB:            db,
		TemplateCache: cache,
	}
}

func (app *App) Render(w http.ResponseWriter, name string, data interface{}) {
	ts, ok := app.TemplateCache[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		log.Printf("Template not found in cache: %s", name)
		return
	}

	// Inject default data like CurrentYear if data is a map
	if dataMap, ok := data.(map[string]interface{}); ok {
		if _, exists := dataMap["CurrentYear"]; !exists {
			dataMap["CurrentYear"] = time.Now().Year()
		}
	}

	// Execute "base.html" which is the layout
	err := ts.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *App) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	posts, err := app.DB.GetAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	aboutContent, err := app.DB.GetSetting("about")
	if err != nil {
		aboutContent = "Welcome! (Edit this in admin)"
	}

	var aboutBuf bytes.Buffer
	if err := goldmark.Convert([]byte(aboutContent), &aboutBuf); err != nil {
		log.Printf("Error rendering about markdown: %v", err)
	}

	data := map[string]interface{}{
		"Posts":     posts,
		"AboutHTML": template.HTML(aboutBuf.String()),
	}

	app.Render(w, "home.html", data)
}