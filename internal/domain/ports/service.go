package ports

import (
	"1337b04rd/internal/domain/models"
	"context"
	"mime/multipart"
)

type PostService interface {
	CreatePost(ctx context.Context, post *models.Post, imageFile multipart.File, imageHeader *multipart.FileHeader) error
	GetPost(ctx context.Context, id int) (*models.Post, error)
	GetPosts(ctx context.Context, limit, offset int, includeArchived bool) ([]*models.Post, error)
	GetPostsByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post, imageFile multipart.File, imageHeader *multipart.FileHeader) error
	DeletePost(ctx context.Context, id int) error
	ArchivePost(ctx context.Context, id int) error
	UnarchivePost(ctx context.Context, id int) error
}

type CommentService interface {
	CreateComment(ctx context.Context, comment *models.Comment, imageFile multipart.File, imageHeader *multipart.FileHeader) error
	GetComment(ctx context.Context, id int) (*models.Comment, error)
	GetCommentsByPost(ctx context.Context, postID int) ([]*models.Comment, error)
	UpdateComment(ctx context.Context, comment *models.Comment, imageFile multipart.File, imageHeader *multipart.FileHeader) error
	DeleteComment(ctx context.Context, id int) error
}

type SessionService interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, id string) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, id string) error
	CleanupExpiredSessions(ctx context.Context) error
}
