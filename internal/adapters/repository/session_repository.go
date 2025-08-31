package repository

import (
	"context"
	"database/sql"
	"errors"
	"1337b04rd/internal/domain/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, name, gender, age, image, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.Name, session.Gender, session.Age, session.Image, session.CreatedAt, session.ExpiresAt,
	)
	
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	query := `
		SELECT id, name, gender, age, image, created_at, expires_at
		FROM sessions WHERE id = $1`
	
	session := &models.Session{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.Name, &session.Gender, &session.Age, &session.Image, &session.CreatedAt, &session.ExpiresAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	return session, nil
}

func (r *SessionRepository) Update(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE sessions SET name = $1, gender = $2, age = $3, image = $4, expires_at = $5 WHERE id = $6`
	
	result, err := r.db.ExecContext(ctx, query,
		session.Name, session.Gender, session.Age, session.Image, session.ExpiresAt, session.ID,
	)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("session not found")
	}
	
	return nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("session not found")
	}
	
	return nil
}

func (r *SessionRepository) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	
	_, err := r.db.ExecContext(ctx, query)
	return err
}
