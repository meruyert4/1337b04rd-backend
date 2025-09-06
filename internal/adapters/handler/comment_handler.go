package handler

import (
	"1337b04rd/internal/adapters/middleware"
	"1337b04rd/internal/adapters/storage"
	"1337b04rd/internal/domain/models"
	"1337b04rd/internal/domain/ports"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CommentHandler struct {
	commentService ports.CommentService
}

func NewCommentHandler(commentService ports.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// convertCommentURLs converts MinIO URLs to backend proxy URLs in a comment
func convertCommentURLs(comment *models.Comment) {
	if comment.ImageURL != "" {
		comment.ImageURL = storage.ConvertMinioURLToProxyURL(comment.ImageURL)
	}
	if comment.AuthorImage != "" {
		comment.AuthorImage = storage.ConvertMinioURLToProxyURL(comment.AuthorImage)
	}
}

// convertCommentsURLs converts MinIO URLs to backend proxy URLs in a slice of comments
func convertCommentsURLs(comments []*models.Comment) {
	for _, comment := range comments {
		convertCommentURLs(comment)
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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

	// Get form data
	postIDStr := r.FormValue("post_id")
	title := r.FormValue("title")
	content := r.FormValue("content")
	replyToCommentIDStr := r.FormValue("reply_to_comment_id")

	if postIDStr == "" || title == "" || content == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Create comment model with session data
	comment := &models.Comment{
		PostID:      postID,
		Title:       title,
		Content:     content,
		AuthorID:    session.ID,
		AuthorName:  session.Name,
		AuthorImage: session.Image,
	}

	// Handle reply to comment if provided
	if replyToCommentIDStr != "" {
		replyToCommentID, err := strconv.Atoi(replyToCommentIDStr)
		if err != nil {
			http.Error(w, "Invalid reply to comment ID", http.StatusBadRequest)
			return
		}
		comment.ReplyToCommentID = &replyToCommentID
	}

	// Get image file if provided
	var imageFile multipart.File
	var imageHeader *multipart.FileHeader
	if file, header, err := r.FormFile("image"); err == nil {
		imageFile = file
		imageHeader = header
		defer imageFile.Close()
	}

	// Create comment
	err = h.commentService.CreateComment(r.Context(), comment, imageFile, imageHeader)
	if err != nil {
		http.Error(w, "Failed to create comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert URLs and return created comment
	convertCommentURLs(comment)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from path parameter
	vars := mux.Vars(r)
	commentIDStr := vars["id"]
	if commentIDStr == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Get comment
	comment, err := h.commentService.GetComment(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to get comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if comment == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Convert URLs and return comment
	convertCommentURLs(comment)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) GetCommentsByPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get comments by post
	comments, err := h.commentService.GetCommentsByPost(r.Context(), postID)
	if err != nil {
		http.Error(w, "Failed to get comments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert URLs and return comments
	convertCommentsURLs(comments)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
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

	// Extract comment ID from path parameter
	vars := mux.Vars(r)
	commentIDStr := vars["id"]
	if commentIDStr == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Get the existing comment to verify ownership
	existingComment, err := h.commentService.GetComment(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to get comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingComment == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Check if the user owns this comment
	if existingComment.AuthorID != session.ID {
		http.Error(w, "Unauthorized: You can only update your own comments", http.StatusForbidden)
		return
	}

	// Get form data
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create comment model with session data
	comment := &models.Comment{
		ID:          commentID,
		Title:       title,
		Content:     content,
		AuthorID:    session.ID,
		AuthorName:  session.Name,
		AuthorImage: session.Image,
	}

	// Get image file if provided
	var imageFile multipart.File
	var imageHeader *multipart.FileHeader
	if file, header, err := r.FormFile("image"); err == nil {
		imageFile = file
		imageHeader = header
		defer imageFile.Close()
	}

	// Update comment
	err = h.commentService.UpdateComment(r.Context(), comment, imageFile, imageHeader)
	if err != nil {
		http.Error(w, "Failed to update comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert URLs and return updated comment
	convertCommentURLs(comment)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract comment ID from path parameter
	vars := mux.Vars(r)
	commentIDStr := vars["id"]
	if commentIDStr == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Get the existing comment to verify ownership
	existingComment, err := h.commentService.GetComment(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to get comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingComment == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Check if the user owns this comment
	if existingComment.AuthorID != session.ID {
		http.Error(w, "Unauthorized: You can only delete your own comments", http.StatusForbidden)
		return
	}

	// Delete comment
	err = h.commentService.DeleteComment(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Failed to delete comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
