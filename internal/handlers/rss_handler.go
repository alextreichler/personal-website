package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

type RSS struct {
	XMLName     xml.Name `xml:"rss"`
	Version     string   `xml:"version,attr"`
	Channel     Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Guid        string `xml:"guid"`
}

func (app *App) RSSFeed(w http.ResponseWriter, r *http.Request) {
	posts, err := app.DB.GetPublishedPosts()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Site configuration (could be moved to settings DB later)
	siteTitle := "Alex Treichler's Blog"
	siteLink := "http://localhost:6060" // TODO: Make this configurable/dynamic based on Host header
	if r.Host != "" {
		siteLink = "http://" + r.Host // Simple protocol assumption, ideally use config
	}
	siteDesc := "Personal website and blog of Alex Treichler."

	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       siteTitle,
			Link:        siteLink,
			Description: siteDesc,
		},
	}

	for _, post := range posts {
		// Create a snippet for description
		desc := post.Content
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}

		item := Item{
			Title:       post.Title,
			Link:        siteLink + "/post/" + post.Slug,
			Description: desc,
			PubDate:     post.CreatedAt.Format(time.RFC1123Z),
			Guid:        siteLink + "/post/" + post.Slug,
		}
		rss.Channel.Items = append(rss.Channel.Items, item)
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(rss); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *App) Sitemap(w http.ResponseWriter, r *http.Request) {
	posts, err := app.DB.GetPublishedPosts()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	w.Write([]byte(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`))

	// Home page
	w.Write([]byte(`
	<url>
		<loc>https://alextreichler.com/</loc>
		<changefreq>daily</changefreq>
		<priority>1.0</priority>
	</url>`))

	for _, post := range posts {
		w.Write([]byte(fmt.Sprintf(`
	<url>
		<loc>https://alextreichler.com/post/%s</loc>
		<lastmod>%s</lastmod>
		<changefreq>monthly</changefreq>
		<priority>0.8</priority>
	</url>`, post.Slug, post.UpdatedAt.Format("2006-01-02"))))
	}

	w.Write([]byte(`</urlset>`))
}
