package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextreichler/personal-website/internal/models"
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
		deleted_at DATETIME,
		status TEXT DEFAULT 'published'
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);

	INSERT OR IGNORE INTO settings (key, value) VALUES ('about', 'Welcome to my new website! Edit this text in the admin dashboard.');
	`

	_, err := d.Conn.Exec(query)
	if err != nil {
		return err
	}

	// Migration for existing databases
	migrations := []string{
		`ALTER TABLE posts ADD COLUMN status TEXT DEFAULT 'published'`,
		`ALTER TABLE posts ADD COLUMN views INTEGER DEFAULT 0`,
	}

	for _, q := range migrations {
		_, err := d.Conn.Exec(q)
		if err != nil {
			// Ignore "duplicate column name" error
			if !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("migration failed: %w", err)
			}
		}
	}

	return nil
}

func (d *Database) GetAllPosts() ([]*models.Post, error) {
	query := `SELECT id, title, slug, content, status, views, created_at, updated_at FROM posts WHERE deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := d.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		if err := rows.Scan(&post.ID, &post.Title, &post.Slug, &post.Content, &post.Status, &post.Views, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (d *Database) GetPublishedPosts() ([]*models.Post, error) {
	query := `SELECT id, title, slug, content, status, created_at, updated_at FROM posts WHERE deleted_at IS NULL AND status = 'published' ORDER BY created_at DESC`
	rows, err := d.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		if err := rows.Scan(&post.ID, &post.Title, &post.Slug, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (d *Database) CreatePost(post *models.Post) error {
	query := `INSERT INTO posts (title, slug, content, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.Conn.Exec(query, post.Title, post.Slug, post.Content, post.Status, post.CreatedAt, post.UpdatedAt)
	return err
}

func (d *Database) GetPostByID(id int) (*models.Post, error) {
	query := `SELECT id, title, slug, content, status, created_at, updated_at FROM posts WHERE id = ? AND deleted_at IS NULL`
	row := d.Conn.QueryRow(query, id)

	post := &models.Post{}
	err := row.Scan(&post.ID, &post.Title, &post.Slug, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (d *Database) GetPostBySlug(slug string) (*models.Post, error) {
	// Increment view count
	_, _ = d.Conn.Exec(`UPDATE posts SET views = views + 1 WHERE slug = ?`, slug)

	query := `SELECT id, title, slug, content, status, views, created_at, updated_at FROM posts WHERE slug = ? AND deleted_at IS NULL AND status = 'published'`
	row := d.Conn.QueryRow(query, slug)

	post := &models.Post{}
	err := row.Scan(&post.ID, &post.Title, &post.Slug, &post.Content, &post.Status, &post.Views, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (d *Database) UpdatePost(post *models.Post) error {
	query := `UPDATE posts SET title = ?, slug = ?, content = ?, status = ?, created_at = ?, updated_at = ? WHERE id = ?`
	_, err := d.Conn.Exec(query, post.Title, post.Slug, post.Content, post.Status, post.CreatedAt, post.UpdatedAt, post.ID)
	return err
}

func (d *Database) DeletePost(id int) error {
	query := `UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.Conn.Exec(query, id)
	return err
}

func (d *Database) GetSetting(key string) (string, error) {
	var value string
	err := d.Conn.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *Database) UpdateSetting(key, value string) error {
	_, err := d.Conn.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}
