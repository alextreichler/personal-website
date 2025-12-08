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
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

// Helper to get a configured Goldmark instance
func getMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
	)
}

func (app *App) ViewPost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/post/")

	post, err := app.DB.GetPostBySlug(slug)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	var safeHTML string
	if post.HTMLContent != "" {
		// Use cached content
		safeHTML = post.HTMLContent
	} else {
		// Fallback: Render on the fly
		var buf bytes.Buffer
		md := getMarkdown()
		if err := md.Convert([]byte(post.Content), &buf); err != nil {
			slog.Error("Error rendering markdown", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		
p := bluemonday.UGCPolicy()
		p.AllowAttrs("style").OnElements("pre", "code", "span")
		safeHTML = p.Sanitize(buf.String())
		
		// Inject loading="lazy" into images
		safeHTML = strings.ReplaceAll(safeHTML, "<img ", "<img loading=\"lazy\" ")
	}

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
	tagsInput := r.FormValue("tags")

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

	// Render Markdown to HTML for caching
	var buf bytes.Buffer
	md := getMarkdown()
	if err := md.Convert([]byte(content), &buf); err == nil {
		p := bluemonday.UGCPolicy()
		p.AllowAttrs("style").OnElements("pre", "code", "span")
		safeHTML := p.Sanitize(buf.String())
		safeHTML = strings.ReplaceAll(safeHTML, "<img ", "<img loading=\"lazy\" ")
		post.HTMLContent = safeHTML
	}

	// We need the ID to set tags, so CreatePost must return the ID or update the struct
	// Assuming CreatePost in db.go uses LastInsertId and likely doesn't return it easily
	// Let's modify CreatePost in db.go to return ID or we assume title+slug is unique enough?
	// Better: Update CreatePost to use LastInsertId. 
	// For now, let's modify db.go CreatePost signature first, OR fetch it back.
	// Let's fetch it back by slug since slug is unique.
	
	if err := app.DB.CreatePost(post); err != nil {
		slog.Error("Error creating post", "error", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}
	
	// Fetch back to get ID
	createdPost, err := app.DB.GetPostBySlug(post.Slug)
	if err == nil {
		tags := strings.Split(tagsInput, ",")
		app.DB.SetPostTags(createdPost.ID, tags)
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
		"Post":       post,
		"TagsString": strings.Join(post.Tags, ", "),
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
	tagsInput := r.FormValue("tags")

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
	
	// Render Markdown to HTML for caching
	var buf bytes.Buffer
	md := getMarkdown()
	if err := md.Convert([]byte(content), &buf); err == nil {
		p := bluemonday.UGCPolicy()
		p.AllowAttrs("style").OnElements("pre", "code", "span")
		safeHTML := p.Sanitize(buf.String())
		safeHTML = strings.ReplaceAll(safeHTML, "<img ", "<img loading=\"lazy\" ")
		post.HTMLContent = safeHTML
	}
	
	// --- End Input Validation ---

	err = app.DB.UpdatePost(post)
	if err != nil {
		slog.Error("Error updating post", "error", err)
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	// Update Tags
	tags := strings.Split(tagsInput, ",")
	if err := app.DB.SetPostTags(post.ID, tags); err != nil {
		slog.Error("Error updating tags", "error", err)
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
