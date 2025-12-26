package handlers

import (
	"bytes"
	"fmt"
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
	"golang.org/x/net/html"
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
		
		// Inject srcset for responsive images
		safeHTML = injectSrcset(safeHTML)
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
	data := make(map[string]interface{})
	app.Render(w, r, "admin_post_new.html", data)
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
		safeHTML = injectSrcset(safeHTML)
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

// injectSrcset parses the HTML and adds srcset attributes to our optimized images
func injectSrcset(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent // Fallback to original if parsing fails
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			var src string
			for _, a := range n.Attr {
				if a.Key == "src" {
					src = a.Val
					break
				}
			}

			// Check if this is one of our optimized images
			// Expected format: .../optimized/optimized_UUID.webp
			if strings.Contains(src, "/optimized/optimized_") && strings.HasSuffix(src, ".webp") {
				base := strings.TrimSuffix(src, ".webp")
				
				// Construct srcset
				// We have: base.webp (1200w), base_800w.webp, base_400w.webp
				srcset := fmt.Sprintf("%s_400w.webp 400w, %s_800w.webp 800w, %s.webp 1200w", base, base, base)
				
				// Add srcset attribute
				n.Attr = append(n.Attr, html.Attribute{Key: "srcset", Val: srcset})
				
				// Add sizes attribute
				sizes := "(max-width: 600px) 400px, (max-width: 900px) 800px, 1200px"
				n.Attr = append(n.Attr, html.Attribute{Key: "sizes", Val: sizes})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return htmlContent
	}
	
	// html.Render wraps content in <html><head></head><body>...</body></html> if it's a full doc,
	// or just nodes. Since we parsed a fragment (likely), html.Parse might add html/body tags.
	// Let's check. html.Parse usually expects a full doc.
	// For fragments, we should traverse the body's children.
	// However, simple hack: render and strip the tags if they were added, or just return the body content.
	// Actually, for post content, it's a fragment. html.Parse will put it in <html><body>...
	
	// Correct approach for fragment:
	// Find <body> and render its children
	var body *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
		}
	}
	findBody(doc)
	
	if body != nil {
		var bodyBuf bytes.Buffer
		for c := body.FirstChild; c != nil; c = c.NextSibling {
			html.Render(&bodyBuf, c)
		}
		return bodyBuf.String()
	}

	return buf.String()
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
		safeHTML = injectSrcset(safeHTML)
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
