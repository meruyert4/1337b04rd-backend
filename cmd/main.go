package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"1337b04rd/config"
	"1337b04rd/internal/adapters/middleware"
	"1337b04rd/internal/app"
	"1337b04rd/pkg/logger"

	"github.com/gorilla/mux"
)

func main() {
	// Setup logger
	logger, logFile := logger.SetupLogger()
	defer logFile.Close()

	// Load configuration
	cfg := config.Load()
	logger.Info("Configuration loaded", "config", cfg)

	// Initialize application
	app, err := app.NewApp(cfg)
	if err != nil {
		logger.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}
	defer app.Close()

	// Setup router
	router := setupRouter(app)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

func setupRouter(app *app.App) *mux.Router {
	router := mux.NewRouter()

	// Add CORS middleware BEFORE any routes
	router.Use(corsMiddleware)

	// Add global OPTIONS handler for CORS preflight
	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origin from environment or use default
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000"
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
		w.WriteHeader(http.StatusOK)
	})

	// Initialize session middleware
	sessionMiddleware := middleware.NewSessionMiddleware(app.SessionService)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Public routes (no session required)
	// Session routes
	sessions := api.PathPrefix("/sessions").Subrouter()
	sessions.HandleFunc("", app.SessionHandler.CreateSession).Methods("POST")
	sessions.HandleFunc("/{id}", app.SessionHandler.GetSession).Methods("GET")
	sessions.HandleFunc("/{id}", app.SessionHandler.UpdateSession).Methods("PUT")
	sessions.HandleFunc("/{id}", app.SessionHandler.DeleteSession).Methods("DELETE")
	sessions.HandleFunc("/cleanup", app.SessionHandler.CleanupExpiredSessions).Methods("POST")
	sessions.HandleFunc("/logout", app.SessionHandler.Logout).Methods("POST")

	// Character routes (Rick and Morty API)
	characters := api.PathPrefix("/characters").Subrouter()
	characters.HandleFunc("/random", app.CharacterHandler.GetRandomCharacter).Methods("GET")
	characters.HandleFunc("", app.CharacterHandler.GetAllCharacters).Methods("GET")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// Protected routes (session required)
	// Post routes (protected)
	posts := api.PathPrefix("/posts").Subrouter()
	posts.Use(sessionMiddleware.ExtractSession)
	posts.HandleFunc("", app.PostHandler.CreatePost).Methods("POST")
	posts.HandleFunc("", app.PostHandler.GetPosts).Methods("GET")
	posts.HandleFunc("/{id:[0-9]+}", app.PostHandler.GetPost).Methods("GET")
	posts.HandleFunc("/{id:[0-9]+}", app.PostHandler.UpdatePost).Methods("PUT")
	posts.HandleFunc("/{id:[0-9]+}", app.PostHandler.DeletePost).Methods("DELETE")
	posts.HandleFunc("/{id:[0-9]+}/archive", app.PostHandler.ArchivePost).Methods("POST")
	posts.HandleFunc("/{id:[0-9]+}/unarchive", app.PostHandler.UnarchivePost).Methods("POST")
	posts.HandleFunc("/author", app.PostHandler.GetPostsByAuthor).Methods("GET")

	// Comment routes (protected)
	comments := api.PathPrefix("/comments").Subrouter()
	comments.Use(sessionMiddleware.ExtractSession)
	comments.HandleFunc("", app.CommentHandler.CreateComment).Methods("POST")
	comments.HandleFunc("/{id:[0-9]+}", app.CommentHandler.GetComment).Methods("GET")
	comments.HandleFunc("/{id:[0-9]+}", app.CommentHandler.UpdateComment).Methods("PUT")
	comments.HandleFunc("/{id:[0-9]+}", app.CommentHandler.DeleteComment).Methods("DELETE")
	comments.HandleFunc("/post", app.CommentHandler.GetCommentsByPost).Methods("GET")

	// Image serving routes
	router.HandleFunc("/images/posts/{filename}", app.Storage.ServePostImageHandler()).Methods("GET")
	router.HandleFunc("/images/comments/{filename}", app.Storage.ServeCommentImageHandler()).Methods("GET")
	router.HandleFunc("/images/avatars/{filename}", app.Storage.ServeAvatarImageHandler()).Methods("GET")
	router.HandleFunc("/images/proxy", app.Storage.ServeImageFromURL()).Methods("GET")

	return router
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origin from environment or use default
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000"
		}

		// Set CORS headers for ALL requests
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For all other requests, continue to the next handler
		next.ServeHTTP(w, r)
	})
}
