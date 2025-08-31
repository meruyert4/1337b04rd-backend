package models

import "time"

type Post struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	AuthorID    string     `json:"author_id"`
	AuthorName  string     `json:"author_name"`
	AuthorImage string     `json:"author_image"`
	ImageURL    string     `json:"image_url"`
	Comments    []*Comment `json:"comments"`
	IsArchive   bool       `json:"is_archive"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
}
