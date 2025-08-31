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

type PostService struct {
	postRepo    ports.PostRepository
	commentRepo ports.CommentRepository
	storage     *storage.MinioClient
}

func NewPostService(postRepo ports.PostRepository, commentRepo ports.CommentRepository, storage *storage.MinioClient) *PostService {
	return &PostService{
		postRepo:    postRepo,
		commentRepo: commentRepo,
		storage:     storage,
	}
}

func (s *PostService) CreatePost(ctx context.Context, post *models.Post, imageFile multipart.File, imageHeader *multipart.FileHeader) error {
	// Set creation time
	post.CreatedAt = time.Now()

	// Handle image upload if provided
	if imageFile != nil && imageHeader != nil {
		imageURL, err := s.storage.UploadPostImage(ctx, imageFile, imageHeader.Filename, imageHeader.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		post.ImageURL = imageURL
	}

	// Create post in database
	err := s.postRepo.Create(ctx, post)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) GetPost(ctx context.Context, id int) (*models.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, errors.New("post not found")
	}

	// Load comments for the post
	comments, err := s.commentRepo.GetByPostID(ctx, id)
	if err != nil {
		return nil, err
	}

	post.Comments = comments
	return post, nil
}

func (s *PostService) GetPosts(ctx context.Context, limit, offset int, includeArchived bool) ([]*models.Post, error) {
	posts, err := s.postRepo.GetAll(ctx, limit, offset, includeArchived)
	if err != nil {
		return nil, err
	}

	// Load comments for each post
	for _, post := range posts {
		comments, err := s.commentRepo.GetByPostID(ctx, post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments
	}

	return posts, nil
}

func (s *PostService) GetPostsByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*models.Post, error) {
	posts, err := s.postRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Load comments for each post
	for _, post := range posts {
		comments, err := s.commentRepo.GetByPostID(ctx, post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments
	}

	return posts, nil
}

func (s *PostService) UpdatePost(ctx context.Context, post *models.Post, imageFile multipart.File, imageHeader *multipart.FileHeader) error {
	// Check if post exists
	existingPost, err := s.postRepo.GetByID(ctx, post.ID)
	if err != nil {
		return err
	}

	if existingPost == nil {
		return errors.New("post not found")
	}

	// Handle image upload if provided
	if imageFile != nil && imageHeader != nil {
		imageURL, err := s.storage.UploadPostImage(ctx, imageFile, imageHeader.Filename, imageHeader.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		post.ImageURL = imageURL
	} else {
		// Keep existing image if no new image provided
		post.ImageURL = existingPost.ImageURL
	}

	// Update post in database
	err = s.postRepo.Update(ctx, post)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) DeletePost(ctx context.Context, id int) error {
	// Check if post exists
	existingPost, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existingPost == nil {
		return errors.New("post not found")
	}

	// Delete post from database
	err = s.postRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) ArchivePost(ctx context.Context, id int) error {
	return s.postRepo.Archive(ctx, id)
}

func (s *PostService) UnarchivePost(ctx context.Context, id int) error {
	return s.postRepo.Unarchive(ctx, id)
}
