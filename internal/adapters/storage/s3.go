package storage

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client        *minio.Client
	avatarBucket  string
	postBucket    string
	commentBucket string
	sessionBucket string // Add session bucket
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Minio client: %w", err)
	}

	// Buckets to create
	buckets := []string{"avatars", "posts", "comments", "sessions"} // Add "sessions" bucket

	for _, bucket := range buckets {
		for attempts := 0; attempts < 3; attempts++ {
			exists, err := client.BucketExists(context.Background(), bucket)
			if err == nil && exists {
				break
			}
			if err == nil {
				err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
				if err != nil {
					return nil, fmt.Errorf("failed to create bucket %s: %w", bucket, err)
				}
				break
			}
			time.Sleep(3 * time.Second)
		}
	}

	return &MinioClient{
		client:        client,
		avatarBucket:  "avatars",
		postBucket:    "posts",
		commentBucket: "comments",
		sessionBucket: "sessions", // Set session bucket
	}, nil
}

// UploadAvatarFromURL uploads an avatar image from a URL to the avatars bucket
func (m *MinioClient) UploadAvatarFromURL(ctx context.Context, imageUrl string) (string, error) {
	data, contentType, err := DownloadFile(ctx, imageUrl)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("invalid content type: %s", contentType)
	}

	objectName := fmt.Sprintf("%d-avatar", time.Now().UnixNano())

	_, err = m.client.PutObject(ctx, m.avatarBucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	return fmt.Sprintf("http://localhost:9001/browser/%s/%s", m.avatarBucket, objectName), nil
}

// UploadCommentImage uploads a comment image from form-data
func (m *MinioClient) UploadCommentImage(ctx context.Context, fileBytes []byte, filename, contentType string) (string, error) {
	objectName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename)

	_, err := m.client.PutObject(ctx, m.commentBucket, objectName, bytes.NewReader(fileBytes), int64(len(fileBytes)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload comment image: %w", err)
	}

	return fmt.Sprintf("/images/comments/%s", objectName), nil
}

// UploadCommentImageFile uploads a comment image from multipart.File
func (m *MinioClient) UploadCommentImageFile(ctx context.Context, file multipart.File, filename, contentType string) (string, error) {
	objectName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename)

	_, err := m.client.PutObject(ctx, m.commentBucket, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload comment image: %w", err)
	}

	return fmt.Sprintf("/images/comments/%s", objectName), nil
}

// SaveSession saves session data to the sessions bucket
func (m *MinioClient) SaveSession(ctx context.Context, sessionID string, sessionData []byte) error {
	objectName := sessionID
	_, err := m.client.PutObject(ctx, m.sessionBucket, objectName, bytes.NewReader(sessionData), int64(len(sessionData)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload session: %w", err)
	}
	return nil
}

func (m *MinioClient) Client() *minio.Client {
	return m.client
}

func (m *MinioClient) UploadImage(ctx context.Context, file multipart.File, filename, contentType string) (string, error) {
	// Generate a unique object name using timestamp and filename
	objectName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename)

	// Upload the image to the 'posts' bucket
	_, err := m.client.PutObject(ctx, m.postBucket, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	// Construct and return the URL of the uploaded image
	imageURL := fmt.Sprintf("/images/posts/%s", objectName)
	return imageURL, nil
}

func (m *MinioClient) GetImage(ctx context.Context, bucket, objectName string) ([]byte, string, error) {
	obj, err := m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("failed to stat object: %w", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj); err != nil {
		return nil, "", fmt.Errorf("failed to read object data: %w", err)
	}

	return buf.Bytes(), stat.ContentType, nil
}

func ServePostImageHandler(storage *MinioClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename") // Go 1.21+
		data, contentType, err := storage.GetImage(r.Context(), "posts", filename)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func ServeCommentImageHandler(storage *MinioClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		data, contentType, err := storage.GetImage(r.Context(), "comments", filename)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// UploadCharacterImageFromURL downloads a character image from a URL and stores it in the sessions bucket
func (m *MinioClient) UploadCharacterImageFromURL(ctx context.Context, imageUrl string, sessionID string) (string, error) {
	data, contentType, err := DownloadFile(ctx, imageUrl)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("invalid content type: %s", contentType)
	}

	// Extract file extension from URL or use default
	fileExt := ".jpg" // Default extension
	if strings.Contains(imageUrl, ".jpeg") {
		fileExt = ".jpeg"
	} else if strings.Contains(imageUrl, ".png") {
		fileExt = ".png"
	} else if strings.Contains(imageUrl, ".gif") {
		fileExt = ".gif"
	}

	// Use session ID and timestamp for uniqueness
	objectName := fmt.Sprintf("%s-character-image%s", sessionID, fileExt)

	_, err = m.client.PutObject(ctx, m.sessionBucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	// Return the MinIO URL for the stored image
	return fmt.Sprintf("http://localhost:9001/browser/%s/%s", m.sessionBucket, objectName), nil
}

func ServeSessionImageHandler(storage *MinioClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		data, contentType, err := storage.GetImage(r.Context(), "sessions", filename)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
