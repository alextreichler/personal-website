package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
	"strconv" // Added this import
	"strings"
	"time"

	"github.com/alextreichler/personal-website/internal/models"
	"github.com/microcosm-cc/bluemonday"
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

	// Sanitize HTML
	p := bluemonday.UGCPolicy()
	safeHTML := p.Sanitize(buf.String())

	// Create description snippet (first 150 chars)
	desc := post.Content
	if len(desc) > 150 {
		desc = desc[:150] + "..."
	}

	data := map[string]interface{}{
		"Post":            post,
		"ContentHTML":     template.HTML(safeHTML),
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

	// --- Input Validation ---
	if strings.TrimSpace(title) == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(content) == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}
	if status != "draft" && status != "published" {
		status = "draft" // Default to draft if invalid status provided
	}

	// Simple slug generation if empty
	if slug == "" {
		slug = slugify(title)
	} else {
		slug = slugify(slug) // Ensure user-provided slug is also slugified
	}
	// --- End Input Validation ---

	now := time.Now()
	post := &models.Post{
		Title:     title,
		Slug:      slug,
		Content:   content,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
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
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-") // Replace spaces with hyphens
	s = strings.Map(func(r rune) rune { // Remove non-alphanumeric characters
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, s)
	s = strings.Trim(s, "-") // Trim leading/trailing hyphens
	// Replace multiple hyphens with a single hyphen
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

func (app *App) AdminEditPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

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
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := app.DB.GetPostByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// --- Input Validation ---
	title := r.FormValue("title")
	content := r.FormValue("content")
	slug := r.FormValue("slug")
	status := r.FormValue("status")

	if strings.TrimSpace(title) == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(content) == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}
	if status != "draft" && status != "published" {
		status = "draft" // Default to draft if invalid status provided
	}

	now := time.Now()

	// Update CreatedAt if publishing a draft
	if post.Status == "draft" && status == "published" {
		post.CreatedAt = now
	}

	post.Title = title
	post.Content = content
	post.Status = status

	if slug == "" {
		post.Slug = slugify(post.Title)
	} else {
		post.Slug = slugify(slug) // Ensure user-provided slug is also slugified
	}
	post.UpdatedAt = now
	// --- End Input Validation ---

	err = app.DB.UpdatePost(post)
	if err != nil {
		slog.Error("Error updating post", "error", err)
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (app *App) AdminDeletePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = app.DB.DeletePost(id)
	if err != nil {
		slog.Error("Error deleting post", "error", err)
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}
