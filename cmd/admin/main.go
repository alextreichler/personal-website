package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/alextreichler/personal-website/internal/auth"
	"github.com/alextreichler/personal-website/internal/config"
	"github.com/alextreichler/personal-website/internal/repository"
)

func main() {
	username := flag.String("user", "", "Username for the new admin")
	password := flag.String("pass", "", "Password for the new admin")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("Usage: go run cmd/admin/main.go -user <username> -pass <password>")
		os.Exit(1)
	}

	// Load Config
	cfg := config.Load()
	// cfg.Validate() // Optional for admin tool, but good practice. 
	// However, admin tool might not need UploadPath created, etc. 
	// Let's minimally use DBPath.

	// Initialize Database
	db, err := repository.NewDatabase(cfg.DBPath)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Conn.Close()

	// Run Migrations (ensure tables exist)
	if err := db.Migrate(); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Create User
	err = auth.CreateUser(db.Conn, *username, *password)
	if err != nil {
		slog.Error("Failed to create user", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created admin user: %s\n", *username)
}