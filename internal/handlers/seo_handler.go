package handlers

import (
		"fmt"
		"net/http"
		"strings"
	)
	
	func (app *App) HealthCheck(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	
	func (app *App) RobotsTXT(w http.ResponseWriter, r *http.Request) {
		// Note: Replace domain with config value if available, or just relative
		// Usually sitemap URL in robots.txt should be absolute.
		// For now hardcoding or using Host header could work, but hardcoding is safer if we know the domain.
		// Since I don't know the user's final domain, I'll use a relative path or just omit the domain if not strict.
		// Better: Use the request Host.
		
		scheme := "https"
	if r.TLS == nil && !strings.Contains(r.Host, "localhost") {
		scheme = "http" 
	}
	
	finalRobots := fmt.Sprintf("User-agent: *\nAllow: /\nDisallow: /admin/\nSitemap: %s://%s/sitemap.xml\n", scheme, r.Host) 

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(finalRobots))
}

func (app *App) Sitemap(w http.ResponseWriter, r *http.Request) {
	// Fetch all posts for sitemap (limit 10000)
	posts, err := app.DB.GetPublishedPosts(10000, 0)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	scheme := "https"
	if r.TLS == nil && !strings.Contains(r.Host, "localhost") {
		scheme = "http"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>` + baseURL + `/</loc>
		<changefreq>daily</changefreq>
		<priority>1.0</priority>
	</url>
	<url>
		<loc>` + baseURL + `/about</loc>
		<changefreq>monthly</changefreq>
		<priority>0.8</priority>
	</url>
`))

	for _, post := range posts {
		url := fmt.Sprintf("%s/post/%s", baseURL, post.Slug)
		date := post.UpdatedAt.Format("2006-01-02")
		w.Write([]byte(fmt.Sprintf(`	<url>
		<loc>%s</loc>
		<lastmod>%s</lastmod>
		<changefreq>weekly</changefreq>
		<priority>0.9</priority>
	</url>
`, url, date)))
	}

	w.Write([]byte(`</urlset>`))
}
