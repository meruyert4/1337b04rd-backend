package models

import "time"

type Comment struct {
	ID               int       `json:"id"`
	PostID           int       `json:"post_id"`
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	AuthorID         string    `json:"author_id"`
	AuthorName       string    `json:"author_name"`
	AuthorImage      string    `json:"author_image"`
	ImageURL         string    `json:"image_url"`
	ReplyToCommentID *int      `json:"reply_to_comment_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
