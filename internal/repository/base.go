package repository

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

type Database struct {
	Conn *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	// Directory creation is handled in config validation

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Performance Optimization: Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable Foreign Keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
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
		html_content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		status TEXT DEFAULT 'published'
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS post_tags (
		post_id INTEGER,
		tag_id INTEGER,
		PRIMARY KEY (post_id, tag_id),
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);

	INSERT OR IGNORE INTO settings (key, value) VALUES ('about', 'Welcome to my new website! Edit this text in the admin dashboard.');

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action TEXT NOT NULL,
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := d.Conn.Exec(query)
	if err != nil {
		return err
	}

	// Migration for existing databases
	migrations := []string{
		`ALTER TABLE posts ADD COLUMN status TEXT DEFAULT 'published'`,
		`ALTER TABLE posts ADD COLUMN views INTEGER DEFAULT 0`,
		`ALTER TABLE posts ADD COLUMN html_content TEXT`,
	}

	// ... existing migration loop ...
	for _, q := range migrations {
		_, err := d.Conn.Exec(q)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			// Log error but continue (or return error if critical)
		}
	}

	return nil
}
