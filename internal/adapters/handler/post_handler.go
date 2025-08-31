package handler

import (
	"1337b04rd/internal/adapters/middleware"
	"1337b04rd/internal/domain/models"
	"1337b04rd/internal/domain/ports"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PostHandler struct {
	postService ports.PostService
}

func NewPostHandler(postService ports.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreatePost called, checking session...")
	// Get session from context
	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		log.Printf("No session found in context, returning 401")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	log.Printf("Session found: %s", session.ID)

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form data
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create post model with session data
	post := &models.Post{
		Title:       title,
		Content:     content,
		AuthorID:    session.ID,
		AuthorName:  session.Name,
		AuthorImage: session.Image,
		IsArchive:   false,
	}

	// Get image file if provided
	var imageFile multipart.File
	var imageHeader *multipart.FileHeader
	if file, header, err := r.FormFile("image"); err == nil {
		imageFile = file
		imageHeader = header
		defer imageFile.Close()
	}

	// Create post
	err = h.postService.CreatePost(r.Context(), post, imageFile, imageHeader)
	if err != nil {
		http.Error(w, "Failed to create post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return created post
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from path parameter
	vars := mux.Vars(r)
	postIDStr := vars["id"]
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get post
	post, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to get post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Return post
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	includeArchivedStr := r.URL.Query().Get("include_archived")

	limit := 10 // default limit
	offset := 0 // default offset
	includeArchived := false

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	if includeArchivedStr == "true" {
		includeArchived = true
	}

	// Get posts
	posts, err := h.postService.GetPosts(r.Context(), limit, offset, includeArchived)
	if err != nil {
		http.Error(w, "Failed to get posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return posts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) GetPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	authorID := r.URL.Query().Get("author_id")
	if authorID == "" {
		http.Error(w, "Author ID is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get posts by author
	posts, err := h.postService.GetPostsByAuthor(r.Context(), authorID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return posts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract post ID from path parameter
	vars := mux.Vars(r)
	postIDStr := vars["id"]
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get the existing post to verify ownership
	existingPost, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to get post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingPost == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check if the user owns this post
	if existingPost.AuthorID != session.ID {
		http.Error(w, "Unauthorized: You can only update your own posts", http.StatusForbidden)
		return
	}

	// Get form data
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create post model with session data
	post := &models.Post{
		ID:         postID,
		Title:      title,
		Content:    content,
		AuthorID:   session.ID,
		AuthorName: session.Name,
	}

	// Get image file if provided
	var imageFile multipart.File
	var imageHeader *multipart.FileHeader
	if file, header, err := r.FormFile("image"); err == nil {
		imageFile = file
		imageHeader = header
		defer imageFile.Close()
	}

	// Update post
	err = h.postService.UpdatePost(r.Context(), post, imageFile, imageHeader)
	if err != nil {
		http.Error(w, "Failed to update post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated post
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract post ID from path parameter
	vars := mux.Vars(r)
	postIDStr := vars["id"]
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get the existing post to verify ownership
	existingPost, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to get post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingPost == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check if the user owns this post
	if existingPost.AuthorID != session.ID {
		http.Error(w, "Unauthorized: You can only delete your own posts", http.StatusForbidden)
		return
	}

	// Delete post
	err = h.postService.DeletePost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to delete post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) ArchivePost(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract post ID from path parameter
	vars := mux.Vars(r)
	postIDStr := vars["id"]
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get the existing post to verify ownership
	existingPost, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to get post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingPost == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check if the user owns this post
	if existingPost.AuthorID != session.ID {
		http.Error(w, "Unauthorized: You can only archive your own posts", http.StatusForbidden)
		return
	}

	// Archive post
	err = h.postService.ArchivePost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to archive post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) UnarchivePost(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract post ID from path parameter
	vars := mux.Vars(r)
	postIDStr := vars["id"]
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get the existing post to verify ownership
	existingPost, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to get post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingPost == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check if the user owns this post
	if existingPost.AuthorID != session.ID {
		http.Error(w, "Unauthorized: You can only unarchive your own posts", http.StatusForbidden)
		return
	}

	// Unarchive post
	err = h.postService.UnarchivePost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to unarchive post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
