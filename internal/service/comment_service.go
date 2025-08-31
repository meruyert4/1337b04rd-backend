package service

import (
	"1337b04rd/internal/adapters/storage"
	"1337b04rd/internal/domain/models"
	"1337b04rd/internal/domain/ports"
	"context"
	"errors"
	"mime/multipart"
	"time"
)

type CommentService struct {
	commentRepo ports.CommentRepository
	storage     *storage.MinioClient
}

func NewCommentService(commentRepo ports.CommentRepository, storage *storage.MinioClient) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		storage:     storage,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, comment *models.Comment, imageFile multipart.File, imageHeader *multipart.FileHeader) error {
	// Set creation time
	comment.CreatedAt = time.Now()

	// Handle image upload if provided
	if imageFile != nil && imageHeader != nil {
		imageURL, err := s.storage.UploadCommentImage(ctx, imageFile, imageHeader.Filename, imageHeader.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		comment.ImageURL = imageURL
	}

	// Create comment in database
	err := s.commentRepo.Create(ctx, comment)
	if err != nil {
		return err
	}

	return nil
}

func (s *CommentService) GetComment(ctx context.Context, id int) (*models.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if comment == nil {
		return nil, errors.New("comment not found")
	}

	return comment, nil
}

func (s *CommentService) GetCommentsByPost(ctx context.Context, postID int) ([]*models.Comment, error) {
	comments, err := s.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *CommentService) UpdateComment(ctx context.Context, comment *models.Comment, imageFile multipart.File, imageHeader *multipart.FileHeader) error {
	// Check if comment exists
	existingComment, err := s.commentRepo.GetByID(ctx, comment.ID)
	if err != nil {
		return err
	}

	if existingComment == nil {
		return errors.New("comment not found")
	}

	// Handle image upload if provided
	if imageFile != nil && imageHeader != nil {
		imageURL, err := s.storage.UploadCommentImage(ctx, imageFile, imageHeader.Filename, imageHeader.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		comment.ImageURL = imageURL
	} else {
		// Keep existing image if no new image provided
		comment.ImageURL = existingComment.ImageURL
	}

	// Update comment in database
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return err
	}

	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id int) error {
	// Check if comment exists
	existingComment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existingComment == nil {
		return errors.New("comment not found")
	}

	// Delete comment from database
	err = s.commentRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
