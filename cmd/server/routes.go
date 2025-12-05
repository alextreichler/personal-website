package main

import (
	"net/http"

	"github.com/alextreichler/personal-website/internal/handlers"
	"github.com/alextreichler/personal-website/internal/middleware"
	"golang.org/x/time/rate"
)
	
	func routes(app *handlers.App) http.Handler {
		mux := http.NewServeMux()
	
		// Rate Limiter: 5 requests per second, burst of 10
		limiter := middleware.NewRateLimiter(rate.Limit(5), 10)
	
		mux.HandleFunc("GET /", app.Home)
		mux.HandleFunc("GET /admin", limiter.Limit(http.HandlerFunc(app.Login)).ServeHTTP)
		mux.HandleFunc("POST /admin", limiter.Limit(http.HandlerFunc(app.LoginPost)).ServeHTTP)
		mux.HandleFunc("GET /logout", app.Logout)
		mux.HandleFunc("GET /post/", app.ViewPost)
				mux.HandleFunc("GET /rss.xml", app.RSSFeed)
				mux.HandleFunc("GET /sitemap.xml", app.Sitemap)
				mux.Handle("GET /metrics", middleware.MetricsHandler())
			
				// Health Check for Kubernetes
				mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("ok"))
				})
			
				// Protected Admin Routes
				mux.HandleFunc("GET /admin/dashboard", middleware.AuthMiddleware(app.AdminDashboard))
				mux.HandleFunc("GET /admin/posts", middleware.AuthMiddleware(app.AdminListPosts))
				mux.HandleFunc("GET /admin/posts/new", middleware.AuthMiddleware(app.AdminNewPost))
				mux.HandleFunc("POST /admin/posts/new", middleware.AuthMiddleware(app.AdminCreatePost))
				mux.HandleFunc("GET /admin/posts/edit", middleware.AuthMiddleware(app.AdminEditPost))
				mux.HandleFunc("POST /admin/posts/edit", middleware.AuthMiddleware(app.AdminUpdatePost))
				mux.HandleFunc("POST /admin/posts/delete", middleware.AuthMiddleware(app.AdminDeletePost))
			
				mux.HandleFunc("GET /admin/about", middleware.AuthMiddleware(app.AdminEditAbout))
				mux.HandleFunc("POST /admin/about", middleware.AuthMiddleware(app.AdminUpdateAbout))
			
				mux.HandleFunc("GET /admin/media", middleware.AuthMiddleware(app.AdminMediaManager))
				mux.HandleFunc("POST /admin/media/upload", middleware.AuthMiddleware(app.AdminUploadImage))
			
				// Static File Server with Cache Headers
				fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(app.Config.StaticPath)))
				mux.Handle("GET /static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Cache static files for 1 year (immutable)
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
					fileServer.ServeHTTP(w, r)
				}))
			
				// Apply Middleware Chain
				// Flow: Request -> Metrics -> Gzip -> Security -> CSRF -> Mux
				return middleware.MetricsMiddleware(
					middleware.GzipMiddleware(
						middleware.SecurityHeadersMiddleware(
							middleware.CSRFMiddleware(mux),
						),
					),
				)
			}
