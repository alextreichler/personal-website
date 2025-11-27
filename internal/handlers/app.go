package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/alextreichler/personal-website/internal/repository"
	"github.com/yuin/goldmark"
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
		"web/template/dashboard.html",
		"web/template/admin_posts.html",
		"web/template/admin_post_new.html",
		"web/template/admin_post_edit.html",
		"web/template/admin_about.html",
		"web/template/post.html",
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



		posts, err := app.DB.GetAllPosts()



		if err != nil {



			log.Printf("Error fetching posts: %v", err)



			http.Error(w, "Internal Server Error", http.StatusInternalServerError)



			return



		}



	



		aboutContent, err := app.DB.GetSetting("about")



		if err != nil {



			// Fallback if setting is missing



			aboutContent = "Welcome! (Edit this in admin)"



		}



	



		// Render Markdown for About section



		var aboutBuf bytes.Buffer



		if err := goldmark.Convert([]byte(aboutContent), &aboutBuf); err != nil {



			log.Printf("Error rendering about markdown: %v", err)



		}



	



		data := map[string]interface{}{



			"CurrentYear":  time.Now().Year(),



			"Posts":        posts,



			"AboutHTML":    template.HTML(aboutBuf.String()),



		}



	



		err = app.Templates.ExecuteTemplate(w, "home.html", data)



	

	if err != nil {

		log.Printf("Error rendering template: %v", err)

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return

	}

}
