package models

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Post struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	HTMLContent string     `json:"html_content"`
	Status      string     `json:"status"` // "draft" or "published"
	Views     int        `json:"views"`
	Tags      []string   `json:"tags"`   // List of tag names
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // For soft delete
}

func (p *Post) ShowUpdated() bool {
	return p.UpdatedAt.Sub(p.CreatedAt) > 5*time.Minute
}

func (p *Post) ReadingTime() string {
	wordCount := len(strings.Fields(p.Content))
	minutes := float64(wordCount) / 200.0
	if minutes < 1 {
		return "1 min read"
	}
	return fmt.Sprintf("%.0f min read", math.Ceil(minutes))
}
