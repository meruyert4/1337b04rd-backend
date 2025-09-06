package repository

import (
	"1337b04rd/internal/domain/models"
	"context"
	"database/sql"
	"errors"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	query := `
		INSERT INTO posts (title, content, author_id, author_name, author_image, image_url, is_archive, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		post.Title, post.Content, post.AuthorID, post.AuthorName, post.AuthorImage,
		post.ImageURL, post.IsArchive, post.CreatedAt, post.ExpiresAt,
	).Scan(&post.ID)

	return err
}

func (r *PostRepository) GetByID(ctx context.Context, id int) (*models.Post, error) {
	query := `
		SELECT id, title, content, author_id, author_name, author_image, image_url, is_archive, created_at, expires_at
		FROM posts WHERE id = $1`

	post := &models.Post{}
	var authorImage sql.NullString
	var imageURL sql.NullString
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &authorImage,
		&imageURL, &post.IsArchive, &post.CreatedAt, &expiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Handle NULL values
	if authorImage.Valid {
		post.AuthorImage = authorImage.String
	}
	if imageURL.Valid {
		post.ImageURL = imageURL.String
	}
	if expiresAt.Valid {
		post.ExpiresAt = expiresAt.Time
	}

	return post, nil
}

func (r *PostRepository) GetAll(ctx context.Context, limit, offset int, includeArchived bool) ([]*models.Post, error) {
	var query string
	var args []interface{}

	if includeArchived {
		query = `
			SELECT id, title, content, author_id, author_name, author_image, image_url, is_archive, created_at, expires_at
			FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	} else {
		query = `
			SELECT id, title, content, author_id, author_name, author_image, image_url, is_archive, created_at, expires_at
			FROM posts WHERE is_archive = false ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var authorImage sql.NullString
		var imageURL sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &authorImage,
			&imageURL, &post.IsArchive, &post.CreatedAt, &expiresAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle NULL values
		if authorImage.Valid {
			post.AuthorImage = authorImage.String
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}
		if expiresAt.Valid {
			post.ExpiresAt = expiresAt.Time
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepository) GetByAuthorID(ctx context.Context, authorID string, limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT id, title, content, author_id, author_name, author_image, image_url, is_archive, created_at, expires_at
		FROM posts WHERE author_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var authorImage sql.NullString
		var imageURL sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &authorImage,
			&imageURL, &post.IsArchive, &post.CreatedAt, &expiresAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle NULL values
		if authorImage.Valid {
			post.AuthorImage = authorImage.String
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}
		if expiresAt.Valid {
			post.ExpiresAt = expiresAt.Time
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepository) Update(ctx context.Context, post *models.Post) error {
	query := `
		UPDATE posts SET title = $1, content = $2, author_id = $3, author_name = $4, 
		image_url = $5, is_archive = $6, expires_at = $7 WHERE id = $8`

	result, err := r.db.ExecContext(ctx, query,
		post.Title, post.Content, post.AuthorID, post.AuthorName,
		post.ImageURL, post.IsArchive, post.ExpiresAt, post.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *PostRepository) Archive(ctx context.Context, id int) error {
	query := `UPDATE posts SET is_archive = true WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *PostRepository) Unarchive(ctx context.Context, id int) error {
	query := `UPDATE posts SET is_archive = false WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("post not found")
	}

	return nil
}
