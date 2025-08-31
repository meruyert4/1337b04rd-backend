package handler

import (
	"1337b04rd/internal/domain/models"
	"1337b04rd/internal/domain/ports"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type SessionHandler struct {
	sessionService ports.SessionService
}

func NewSessionHandler(sessionService ports.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	log.Printf("Creating new session...")

	// Create session model (all character data will be populated automatically from Rick and Morty API)
	session := &models.Session{}

	// Create session (this will automatically fetch character data and store image)
	err := h.sessionService.CreateSession(r.Context(), session)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Session created successfully with ID: %s", session.ID)

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	// Return created session
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path parameter
	vars := mux.Vars(r)
	sessionID := vars["id"]
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Get session
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Failed to get session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Return session
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (h *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path parameter
	vars := mux.Vars(r)
	sessionID := vars["id"]
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Get form data
	name := r.FormValue("name")
	gender := r.FormValue("gender")
	age := r.FormValue("age")

	// Validate required fields
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Validate gender enum
	if gender != "" && gender != "male" && gender != "female" {
		http.Error(w, "Gender must be 'male' or 'female'", http.StatusBadRequest)
		return
	}

	// Get the existing session to preserve unchanged fields
	existingSession, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Failed to get session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingSession == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update only the provided fields
	if name != "" {
		existingSession.Name = name
	}
	if gender != "" {
		existingSession.Gender = gender
	}
	if age != "" {
		existingSession.Age = age
	}

	// Update session
	err = h.sessionService.UpdateSession(r.Context(), existingSession)
	if err != nil {
		http.Error(w, "Failed to update session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated session
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingSession)
}

func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path parameter
	vars := mux.Vars(r)
	sessionID := vars["id"]
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Delete session
	err := h.sessionService.DeleteSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Failed to delete session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) CleanupExpiredSessions(w http.ResponseWriter, r *http.Request) {
	// Cleanup expired sessions
	err := h.sessionService.CleanupExpiredSessions(r.Context())
	if err != nil {
		http.Error(w, "Failed to cleanup expired sessions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Expired sessions cleaned up successfully"))
}

func (h *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete the cookie
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
