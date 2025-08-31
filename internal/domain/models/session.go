package models

import (
	"time"
)

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"`
	Age       string    `json:"age"` // Using string since Rick and Morty API doesn't provide exact age
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
