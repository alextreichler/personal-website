package repository

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	Conn *sql.DB
}

func NewDatabase() (*Database, error) {
	// Ensure the data directory exists
	dbPath := "./data"
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}

	file := filepath.Join(dbPath, "site.db")
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{Conn: db}, nil
}

func (d *Database) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME
	);
	`

	_, err := d.Conn.Exec(query)
	return err
}
