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
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Minio client: %w", err)
	}

	// Create three buckets: avatars, posts, comments
	buckets := []string{"avatars", "posts", "comments"}

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

	// Makes buckets public
	for _, bucket := range buckets {
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::%s/*"
				}
			]
		}`, bucket)

		err := client.SetBucketPolicy(context.Background(), bucket, policy)
		if err != nil {
			fmt.Printf("Warning: Failed to set public policy for bucket %s: %v\n", bucket, err)
		}
	}

	return &MinioClient{
		client:        client,
		avatarBucket:  "avatars",
		postBucket:    "posts",
		commentBucket: "comments",
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

	// Return the MinIO URL for the stored image
	return fmt.Sprintf("http://localhost:9000/%s/%s", m.avatarBucket, objectName), nil
}

// UploadPostImage uploads a post image from multipart.File
func (m *MinioClient) UploadPostImage(ctx context.Context, file multipart.File, filename, contentType string) (string, error) {
	objectName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename)

	_, err := m.client.PutObject(ctx, m.postBucket, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload post image: %w", err)
	}

	// Return the MinIO URL for the stored image
	return fmt.Sprintf("http://localhost:9000/%s/%s", m.postBucket, objectName), nil
}

// UploadCommentImage uploads a comment image from multipart.File
func (m *MinioClient) UploadCommentImage(ctx context.Context, file multipart.File, filename, contentType string) (string, error) {
	objectName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename)

	_, err := m.client.PutObject(ctx, m.commentBucket, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload comment image: %w", err)
	}

	// Return the MinIO URL for the stored image
	return fmt.Sprintf("http://localhost:9000/%s/%s", m.commentBucket, objectName), nil
}

// UploadCommentImageBytes uploads a comment image from bytes
func (m *MinioClient) UploadCommentImageBytes(ctx context.Context, fileBytes []byte, filename, contentType string) (string, error) {
	objectName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename)

	_, err := m.client.PutObject(ctx, m.commentBucket, objectName, bytes.NewReader(fileBytes), int64(len(fileBytes)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload comment image: %w", err)
	}

	// Return the MinIO URL for the stored image
	return fmt.Sprintf("http://localhost:9000/%s/%s", m.commentBucket, objectName), nil
}

// GetImage retrieves an image from any bucket
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

// ServeAvatarImageHandler serves avatar images from the avatars bucket
func (m *MinioClient) ServeAvatarImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		data, contentType, err := m.GetImage(r.Context(), m.avatarBucket, filename)
		if err != nil {
			http.Error(w, "Avatar image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// ServePostImageHandler serves post images from the posts bucket
func (m *MinioClient) ServePostImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		data, contentType, err := m.GetImage(r.Context(), m.postBucket, filename)
		if err != nil {
			http.Error(w, "Post image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// ServeImageFromURL serves any image from a MinIO URL
func (m *MinioClient) ServeImageFromURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the MinIO URL from query parameter
		imageURL := r.URL.Query().Get("url")
		if imageURL == "" {
			http.Error(w, "Missing image URL", http.StatusBadRequest)
			return
		}

		// Extract bucket and filename from MinIO URL
		// Format: http://localhost:9000/bucket/filename
		if !strings.HasPrefix(imageURL, "http://localhost:9000/") {
			http.Error(w, "Invalid MinIO URL", http.StatusBadRequest)
			return
		}

		path := strings.TrimPrefix(imageURL, "http://localhost:9000/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		bucket := parts[0]
		filename := parts[1]

		// Get image from MinIO
		data, contentType, err := m.GetImage(r.Context(), bucket, filename)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// ServeCommentImageHandler serves comment images from the comments bucket
func (m *MinioClient) ServeCommentImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		data, contentType, err := m.GetImage(r.Context(), m.commentBucket, filename)
		if err != nil {
			http.Error(w, "Comment image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// UploadCharacterImageFromURL downloads a character image from a URL and stores it in the avatars bucket
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

	_, err = m.client.PutObject(ctx, m.avatarBucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	// Return the MinIO URL for the stored image
	return fmt.Sprintf("http://localhost:9000/%s/%s", m.avatarBucket, objectName), nil
}

// DeleteImage deletes an image from any bucket
func (m *MinioClient) DeleteImage(ctx context.Context, bucket, objectName string) error {
	return m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

// GetBucketName returns the bucket name for a given type
func (m *MinioClient) GetBucketName(imageType string) string {
	switch imageType {
	case "avatar":
		return m.avatarBucket
	case "post":
		return m.postBucket
	case "comment":
		return m.commentBucket
	default:
		return m.avatarBucket
	}
}

// ConvertMinioURLToProxyURL converts a MinIO URL to use the backend proxy
func ConvertMinioURLToProxyURL(minioURL string) string {
	// Convert: http://localhost:9000/bucket/filename
	// To:      http://localhost:8080/images/proxy?url=http://localhost:9000/bucket/filename
	if strings.HasPrefix(minioURL, "http://localhost:9000/") {
		return fmt.Sprintf("http://localhost:8080/images/proxy?url=%s", minioURL)
	}
	return minioURL // Return original if not a MinIO URL
}
