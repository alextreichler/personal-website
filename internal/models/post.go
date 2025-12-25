package models

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type DashboardStats struct {
	TotalPosts     int
	PublishedPosts int
	DraftPosts     int
	TotalViews     int
	TopPosts       []*Post
}

type Post struct {
	ID          int
	Title       string
	Slug        string
	Content     string
	HTMLContent string
	Status      string // "draft", "published"
	Tags        []string
	Views       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
