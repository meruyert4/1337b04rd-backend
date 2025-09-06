package repository

import (
	"1337b04rd/internal/domain/models"
	"context"
	"database/sql"
	"errors"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (post_id, title, content, author_id, author_name, author_image, image_url, reply_to_comment_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		comment.PostID, comment.Title, comment.Content, comment.AuthorID,
		comment.AuthorName, comment.AuthorImage, comment.ImageURL, comment.ReplyToCommentID, comment.CreatedAt,
	).Scan(&comment.ID)

	return err
}

func (r *CommentRepository) GetByID(ctx context.Context, id int) (*models.Comment, error) {
	query := `
		SELECT id, post_id, title, content, author_id, author_name, author_image, image_url, reply_to_comment_id, created_at
		FROM comments WHERE id = $1`

	comment := &models.Comment{}
	var authorImage sql.NullString
	var imageURL sql.NullString
	var replyToCommentID sql.NullInt64
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID, &comment.PostID, &comment.Title, &comment.Content, &comment.AuthorID,
		&comment.AuthorName, &authorImage, &imageURL, &replyToCommentID, &comment.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Handle NULL values
	if authorImage.Valid {
		comment.AuthorImage = authorImage.String
	}
	if imageURL.Valid {
		comment.ImageURL = imageURL.String
	}
	if replyToCommentID.Valid {
		replyID := int(replyToCommentID.Int64)
		comment.ReplyToCommentID = &replyID
	}

	return comment, nil
}

func (r *CommentRepository) GetByPostID(ctx context.Context, postID int) ([]*models.Comment, error) {
	query := `
		SELECT id, post_id, title, content, author_id, author_name, author_image, image_url, reply_to_comment_id, created_at
		FROM comments WHERE post_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		var authorImage sql.NullString
		var imageURL sql.NullString
		var replyToCommentID sql.NullInt64
		
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.Title, &comment.Content, &comment.AuthorID,
			&comment.AuthorName, &authorImage, &imageURL, &replyToCommentID, &comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Handle NULL values
		if authorImage.Valid {
			comment.AuthorImage = authorImage.String
		}
		if imageURL.Valid {
			comment.ImageURL = imageURL.String
		}
		if replyToCommentID.Valid {
			replyID := int(replyToCommentID.Int64)
			comment.ReplyToCommentID = &replyID
		}
		
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *models.Comment) error {
	query := `
		UPDATE comments SET title = $1, content = $2, author_id = $3, author_name = $4, 
		image_url = $5, reply_to_comment_id = $6 WHERE id = $7`

	result, err := r.db.ExecContext(ctx, query,
		comment.Title, comment.Content, comment.AuthorID, comment.AuthorName,
		comment.ImageURL, comment.ReplyToCommentID, comment.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("comment not found")
	}

	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("comment not found")
	}

	return nil
}
