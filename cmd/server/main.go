package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alextreichler/personal-website/internal/auth"
	"github.com/alextreichler/personal-website/internal/config"
	"github.com/alextreichler/personal-website/internal/handlers"
	"github.com/alextreichler/personal-website/internal/repository"
)

func main() {
	// Setup structured logger (JSON is better for K8s/Log aggregators)
	// We'll use Text for dev, but K8s envs often prefer JSON.
	// You might want to switch this based on an env var like APP_ENV=production
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load Configuration
	cfg := config.Load()
	cfg.Validate()

	// Set global session secret
	auth.SecretKey = []byte(cfg.SessionSecret)

	// Handle "migrate" subcommand for InitContainers
	migrateOnly := flag.Bool("migrate", false, "Run database migrations and exit")
	flag.Parse()

	if *migrateOnly {
		logger.Info("Starting database migration...")
		db, err := repository.NewDatabase(cfg.DBPath)
		if err != nil {
			logger.Error("Failed to connect to database for migration", "error", err)
			os.Exit(1)
		}
		if err := db.Migrate(); err != nil {
			logger.Error("Migration failed", "error", err)
			os.Exit(1)
		}
		logger.Info("Migration completed successfully.")
		return
	}

	// Initialize Database
	db, err := repository.NewDatabase(cfg.DBPath)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Conn.Close()

	// Run Migrations (optional on startup, but kept for standalone safety)
	if err := db.Migrate(); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database initialized")

	// Initialize Application Handlers
	app := handlers.NewApp(db, cfg)

	// Initialize Server
	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: routes(app),
	}

	// Graceful Shutdown Channel
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("Server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for signal
	<-done
	logger.Info("Server stopped")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("Server exited properly")
}
