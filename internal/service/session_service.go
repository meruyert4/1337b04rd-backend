package service

import (
	"1337b04rd/internal/adapters/externalapi"
	"1337b04rd/internal/adapters/storage"
	"1337b04rd/internal/domain/models"
	"1337b04rd/internal/domain/ports"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"
)

type SessionService struct {
	sessionRepo        ports.SessionRepository
	rickAndMortyClient *externalapi.RickAndMortyClient
	storage            *storage.MinioClient
}

func NewSessionService(sessionRepo ports.SessionRepository, storage *storage.MinioClient) *SessionService {
	return &SessionService{
		sessionRepo:        sessionRepo,
		rickAndMortyClient: externalapi.NewRickAndMortyClient(),
		storage:            storage,
	}
}

// generateSessionID creates a unique session ID using timestamp and random bytes
func generateSessionID() string {
	// Use timestamp for uniqueness
	timestamp := time.Now().UnixNano()

	// Add random bytes for additional uniqueness
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)

	// Combine timestamp and random bytes
	return fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(randomBytes))
}

func (s *SessionService) CreateSession(ctx context.Context, session *models.Session) error {
	// Generate unique session ID
	session.ID = generateSessionID()

	// Set creation time
	session.CreatedAt = time.Now()

	// Set default expiration time (24 hours from now)
	session.ExpiresAt = time.Now().Add(24 * time.Hour)

	// Try to get a random character from Rick and Morty API for session creation
	character, err := s.rickAndMortyClient.FetchRandomCharacter()
	if err != nil {
		// Log the error but continue with default values
		log.Printf("Warning: Failed to fetch character from Rick and Morty API: %v", err)
		// Set default values
		session.Name = "Anonymous User"
		session.Gender = "Unknown"
		session.Age = "Unknown"
		session.Image = ""
	} else {
		// Populate session with character data
		session.Name = character.Name
		session.Gender = character.Gender
		session.Age = character.GetAge()

		// Try to download and store character image in MinIO
		if character.Image != "" {
			// Try MinIO upload, but don't fail if it doesn't work
			minioImageURL, err := s.storage.UploadCharacterImageFromURL(ctx, character.Image, session.ID)
			if err != nil {
				// Log error but continue with original image URL as fallback
				log.Printf("Warning: Failed to upload character image to MinIO: %v", err)
				log.Printf("Continuing with original image URL: %s", character.Image)
				session.Image = character.Image
			} else {
				log.Printf("Successfully uploaded image to MinIO: %s", minioImageURL)
				session.Image = minioImageURL
			}
		} else {
			session.Image = ""
		}
	}

	// Ensure we have at least some values
	if session.Name == "" {
		session.Name = "Anonymous User"
	}
	if session.Gender == "" {
		session.Gender = "Unknown"
	}
	if session.Age == "" {
		session.Age = "Unknown"
	}

	// Create session in database
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to create session in database: %w", err)
	}

	log.Printf("Session created successfully with ID: %s, Name: %s", session.ID, session.Name)
	return nil
}

func (s *SessionService) GetSession(ctx context.Context, id string) (*models.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, errors.New("session not found")
	}

	// Check if session is expired
	if !session.ExpiresAt.IsZero() && time.Now().After(session.ExpiresAt) {
		// Delete expired session
		s.sessionRepo.Delete(ctx, id)
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (s *SessionService) UpdateSession(ctx context.Context, session *models.Session) error {
	// Check if session exists
	existingSession, err := s.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return err
	}

	if existingSession == nil {
		return errors.New("session not found")
	}

	// Update session in database
	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionService) DeleteSession(ctx context.Context, id string) error {
	// Check if session exists
	existingSession, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existingSession == nil {
		return errors.New("session not found")
	}

	// Delete session from database
	err = s.sessionRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.CleanupExpired(ctx)
}
