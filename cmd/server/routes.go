package main

import (
	"net/http"

	"github.com/alextreichler/personal-website/internal/handlers"
)

func routes(app *handlers.App) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", app.Home)
	mux.HandleFunc("GET /login", app.Login)
	mux.HandleFunc("POST /login", app.LoginPost)
	mux.HandleFunc("GET /logout", app.Logout)
	
	// Add static file server later
	// mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	return mux
}
