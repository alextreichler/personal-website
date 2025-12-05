package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/alextreichler/personal-website/internal/models"
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
		// Populate tags
		tags, err := d.GetTagsForPost(post.ID)
		if err == nil {
			post.Tags = tags
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
		// Populate tags
		tags, err := d.GetTagsForPost(post.ID)
		if err == nil {
			post.Tags = tags
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
	
	tags, err := d.GetTagsForPost(post.ID)
	if err == nil {
		post.Tags = tags
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

	tags, err := d.GetTagsForPost(post.ID)
	if err == nil {
		post.Tags = tags
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

func (d *Database) GetTagsForPost(postID int) ([]string, error) {
	query := `
		SELECT t.name 
		FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = ?
		ORDER BY t.name ASC
	`
	rows, err := d.Conn.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, nil
}

func (d *Database) SetPostTags(postID int, tags []string) error {
	tx, err := d.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Clear existing tags for this post
	if _, err := tx.Exec("DELETE FROM post_tags WHERE post_id = ?", postID); err != nil {
		return err
	}

	// 2. Insert new tags and links
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		tagName = strings.ToLower(tagName) // Normalize to lowercase

		// Ensure tag exists
		var tagID int
		err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
		if err == sql.ErrNoRows {
			res, err := tx.Exec("INSERT INTO tags (name) VALUES (?)", tagName)
			if err != nil {
				return err
			}
			id, _ := res.LastInsertId()
			tagID = int(id)
		} else if err != nil {
			return err
		}

		// Link tag to post
		if _, err := tx.Exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", postID, tagID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *Database) GetPostsByTag(tagName string) ([]*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.content, p.status, p.created_at, p.updated_at 
		FROM posts p
		JOIN post_tags pt ON p.id = pt.post_id
		JOIN tags t ON pt.tag_id = t.id
		WHERE t.name = ? AND p.deleted_at IS NULL AND p.status = 'published'
		ORDER BY p.created_at DESC
	`
	rows, err := d.Conn.Query(query, tagName)
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
