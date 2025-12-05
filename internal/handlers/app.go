package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alextreichler/personal-website/internal/auth"
	"github.com/alextreichler/personal-website/internal/config"
	"github.com/alextreichler/personal-website/internal/middleware"
	"github.com/alextreichler/personal-website/internal/repository"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

type App struct {
	DB            *repository.Database
	TemplateCache map[string]*template.Template
	Config        *config.Config
}

func NewApp(db *repository.Database, cfg *config.Config) *App {
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
		"admin_media.html",
		"post.html",
		"error.html",
		// Add other templates here as they are created
	}

	for _, page := range pages {
		name := page
		
		// Corrected: Ensure paths are explicit relative to the project root.
		// The error suggests "web/template/" might be prefixed again internally.
		// By providing the full relative paths directly, we avoid filepath.Join's
		// potential for unexpected behavior with ParseFiles in this context.
		// TODO: Use cfg.StaticPath logic if templates move, but they are in template/ not static/
		ts, err := template.ParseFiles("web/template/base.html", "web/template/"+page)
		if err != nil {
			slog.Error("Error parsing template", "name", name, "error", err)
			os.Exit(1)
		}

		cache[name] = ts
	}

	return &App{
		DB:            db,
		TemplateCache: cache,
		Config:        cfg,
	}
}

func (app *App) Render(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	ts, ok := app.TemplateCache[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		slog.Error("Template not found in cache", "name", name)
		return
	}

	// Inject default data like CurrentYear if data is a map
	if dataMap, ok := data.(map[string]interface{}); ok {
		if _, exists := dataMap["CurrentYear"]; !exists {
			dataMap["CurrentYear"] = time.Now().Year()
		}

		// Inject login status
		cookie, err := r.Cookie(app.Config.SessionCookie)
		if err == nil && cookie.Value != "" {
			username, err := auth.Verify(cookie.Value)
			if err == nil {
				dataMap["IsLoggedIn"] = true
				dataMap["Username"] = username
			} else {
				dataMap["IsLoggedIn"] = false
				dataMap["Username"] = ""
			}
		} else {
			dataMap["IsLoggedIn"] = false
			dataMap["Username"] = ""
		}

		// Inject CSRF token
		dataMap["CSRFToken"] = middleware.GetCSRFToken(r)
	}

	// Execute "base.html" which is the layout
	err := ts.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		slog.Error("Error rendering template", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *App) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.NotFound(w, r)
		return
	}

	posts, err := app.DB.GetPublishedPosts()
	if err != nil {
		slog.Error("Error fetching posts", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	aboutContent, err := app.DB.GetSetting("about")
	if err != nil {
		aboutContent = "Welcome! (Edit this in admin)"
	}

	var aboutBuf bytes.Buffer
	if err := goldmark.Convert([]byte(aboutContent), &aboutBuf); err != nil {
		slog.Error("Error rendering about markdown", "error", err)
	}

	// Sanitize HTML
	p := bluemonday.UGCPolicy()
	safeAboutHTML := p.Sanitize(aboutBuf.String())

	data := map[string]interface{}{
		"Posts":           posts,
		"AboutHTML":       template.HTML(safeAboutHTML),
		"PageTitle":       "Home",
		"MetaDescription": "Welcome to the personal website and blog of Alex Treichler. Read my latest thoughts on technology and more.",
	}

	app.Render(w, r, "home.html", data)
}

func (app *App) RenderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.WriteHeader(status)
	data := map[string]interface{}{
		"StatusCode": status,
		"Message":    message,
		"PageTitle":  "Error",
	}
	app.Render(w, r, "error.html", data)
}

func (app *App) NotFound(w http.ResponseWriter, r *http.Request) {
	app.RenderError(w, r, http.StatusNotFound, "Page Not Found")
}


	