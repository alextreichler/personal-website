package models

import "time"

type Post struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Slug      string     `json:"slug"`
	Content   string     `json:"content"`
	Status    string     `json:"status"` // "draft" or "published"
	Views     int        `json:"views"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // For soft delete
}
