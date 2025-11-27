package main

import (
	"net/http"

	"github.com/alextreichler/personal-website/internal/handlers"
	"github.com/alextreichler/personal-website/internal/middleware"
)

func routes(app *handlers.App) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", app.Home)
	mux.HandleFunc("GET /admin", app.Login)
	mux.HandleFunc("POST /admin", app.LoginPost)
	mux.HandleFunc("GET /logout", app.Logout)
	mux.HandleFunc("GET /post/", app.ViewPost)
	mux.HandleFunc("GET /rss.xml", app.RSSFeed)

	// Protected Admin Routes
	mux.HandleFunc("GET /admin/dashboard", middleware.AuthMiddleware(app.AdminDashboard))
	mux.HandleFunc("GET /admin/posts", middleware.AuthMiddleware(app.AdminListPosts))
	mux.HandleFunc("GET /admin/posts/new", middleware.AuthMiddleware(app.AdminNewPost))
	mux.HandleFunc("POST /admin/posts/new", middleware.AuthMiddleware(app.AdminCreatePost))
	mux.HandleFunc("GET /admin/posts/edit", middleware.AuthMiddleware(app.AdminEditPost))
	mux.HandleFunc("POST /admin/posts/edit", middleware.AuthMiddleware(app.AdminUpdatePost))
	mux.HandleFunc("GET /admin/posts/delete", middleware.AuthMiddleware(app.AdminDeletePost))

	mux.HandleFunc("GET /admin/about", middleware.AuthMiddleware(app.AdminEditAbout))
	mux.HandleFunc("POST /admin/about", middleware.AuthMiddleware(app.AdminUpdateAbout))

	mux.HandleFunc("GET /admin/media", middleware.AuthMiddleware(app.AdminMediaManager))
	mux.HandleFunc("POST /admin/media/upload", middleware.AuthMiddleware(app.AdminUploadImage))
	
	// Static File Server
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	return mux
}
