package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/alextreichler/personal-website/internal/handlers"
	"github.com/alextreichler/personal-website/internal/repository"
)

func main() {
	// Setup structured logger (TextHandler for readability)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize Database
	db, err := repository.NewDatabase()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Conn.Close()

	// Run Migrations
	if err := db.Migrate(); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database initialized and migrated successfully")

	// Initialize Application Handlers
	app := handlers.NewApp(db)

	// Initialize Server
	port := ":6060"
	srv := &http.Server{
		Addr:    port,
		Handler: routes(app),
	}

	logger.Info("Server starting", "port", port)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
