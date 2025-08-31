package middleware

import (
	"1337b04rd/internal/domain/models"
	"1337b04rd/internal/domain/ports"
	"context"
	"log"
	"net/http"
)

type SessionMiddleware struct {
	sessionService ports.SessionService
}

func NewSessionMiddleware(sessionService ports.SessionService) *SessionMiddleware {
	return &SessionMiddleware{
		sessionService: sessionService,
	}
}

// SessionContextKey is the key used to store session in request context
type SessionContextKey string

const (
	SessionContextKeyValue SessionContextKey = "session"
)

// ExtractSession extracts session from cookies and adds it to request context
func (m *SessionMiddleware) ExtractSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from cookie
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			// No session cookie, continue without session
			log.Printf("No session cookie found for request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("Found session cookie: %s", cookie.Value)

		// Get session from database
		session, err := m.sessionService.GetSession(r.Context(), cookie.Value)
		if err != nil || session == nil {
			// Invalid session, continue without session
			log.Printf("Invalid session: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("Session extracted successfully: %s", session.ID)

		// Add session to request context
		ctx := context.WithValue(r.Context(), SessionContextKeyValue, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSessionFromContext extracts session from request context
func GetSessionFromContext(ctx context.Context) *models.Session {
	if session, ok := ctx.Value(SessionContextKeyValue).(*models.Session); ok {
		return session
	}
	return nil
}
