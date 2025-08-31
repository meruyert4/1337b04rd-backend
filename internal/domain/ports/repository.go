package ports

import (
	"context"
	"1337b04rd/internal/domain/models"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id int) (*models.Post, error)
	GetAll(ctx context.Context, limit, offset int, includeArchived bool) ([]*models.Post, error)
	GetByAuthorID(ctx context.Context, authorID string, limit, offset int) ([]*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id int) error
	Archive(ctx context.Context, id int) error
	Unarchive(ctx context.Context, id int) error
}

type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	GetByID(ctx context.Context, id int) (*models.Comment, error)
	GetByPostID(ctx context.Context, postID int) ([]*models.Comment, error)
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, id int) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id string) (*models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, id string) error
	CleanupExpired(ctx context.Context) error
}
