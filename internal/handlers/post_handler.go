package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/alextreichler/personal-website/internal/models"
	"github.com/yuin/goldmark"
)

func (app *App) ViewPost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/post/")
	
	post, err := app.DB.GetPostBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(post.Content), &buf); err != nil {
		slog.Error("Error rendering markdown", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create description snippet (first 150 chars)
	desc := post.Content
	if len(desc) > 150 {
		desc = desc[:150] + "..."
	}

	data := map[string]interface{}{
		"Post":            post,
		"ContentHTML":     template.HTML(buf.String()),
		"PageTitle":       post.Title,
		"MetaDescription": desc,
	}

	app.Render(w, r, "post.html", data)
}

func (app *App) AdminListPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := app.DB.GetAllPosts()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts": posts,
	}

	app.Render(w, r, "admin_posts.html", data)
}

func (app *App) AdminNewPost(w http.ResponseWriter, r *http.Request) {
	app.Render(w, r, "admin_post_new.html", nil)
}

func (app *App) AdminCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	slug := r.FormValue("slug")
	status := r.FormValue("status")
	if status == "" {
		status = "draft" // Default to draft if not specified
	}

	// Simple slug generation if empty
	if slug == "" {
		slug = slugify(title)
	}

	post := &models.Post{
		Title:     title,
		Slug:      slug,
		Content:   content,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Views:     0,
	}

	err = app.DB.CreatePost(post)
	if err != nil {
		slog.Error("Error creating post", "error", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func slugify(s string) string {
	// Very basic slugify: lower case, replace spaces with hyphens
	// In a real app, use a regex to remove non-alphanumeric chars
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

func (app *App) AdminEditPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id := 0
	fmt.Sscanf(idStr, "%d", &id)

	post, err := app.DB.GetPostByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Post": post,
	}

	app.Render(w, r, "admin_post_edit.html", data)
}

func (app *App) AdminUpdatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id := 0
	fmt.Sscanf(idStr, "%d", &id)

	post, err := app.DB.GetPostByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	post.Title = r.FormValue("title")
	post.Slug = r.FormValue("slug")
	post.Status = r.FormValue("status")
	if post.Slug == "" {
		post.Slug = slugify(post.Title)
	}
	post.Content = r.FormValue("content")
	post.UpdatedAt = time.Now()

	err = app.DB.UpdatePost(post)
	if err != nil {
		slog.Error("Error updating post", "error", err)
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (app *App) AdminDeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id := 0
	fmt.Sscanf(idStr, "%d", &id)

	err := app.DB.DeletePost(id)
	if err != nil {
		slog.Error("Error deleting post", "error", err)
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}
