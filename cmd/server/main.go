package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alextreichler/personal-website/internal/handlers"
	"github.com/alextreichler/personal-website/internal/repository"
)

func main() {
	// Initialize Database
	db, err := repository.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Conn.Close()

	// Run Migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("Database initialized and migrated successfully.")

	// Initialize Application Handlers
	app := handlers.NewApp(db)

	// Initialize Server
	srv := &http.Server{
		Addr:    ":6060",
		Handler: routes(app),
	}

	fmt.Printf("Server starting on port %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
