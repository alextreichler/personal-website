package config

import (
	"log/slog"
	"os"
	"path/filepath"
)

type Config struct {
	Port          string
	DBPath        string
	UploadPath    string
	StaticPath    string
	SessionSecret string
	SessionCookie string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", ":6060"),
		DBPath:        getEnv("DB_PATH", "./data/site.db"),
		UploadPath:    getEnv("UPLOAD_PATH", "web/static/uploads"),
		StaticPath:    getEnv("STATIC_PATH", "./web/static"),
		SessionSecret: getEnv("SESSION_SECRET", "default-insecure-secret-change-me"), // Provide default for dev, warn in prod
		SessionCookie: getEnv("SESSION_COOKIE_NAME", "admin_session"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Validate checks for critical configuration issues
func (c *Config) Validate() {
	if c.SessionSecret == "default-insecure-secret-change-me" {
		slog.Warn("Using default insecure SESSION_SECRET. Please set this environment variable in production.")
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(c.UploadPath, 0755); err != nil {
		slog.Error("Failed to create upload directory", "path", c.UploadPath, "error", err)
		os.Exit(1)
	}
	
	// Ensure db directory exists
	dbDir := filepath.Dir(c.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("Failed to create database directory", "path", dbDir, "error", err)
		os.Exit(1)
	}
}
