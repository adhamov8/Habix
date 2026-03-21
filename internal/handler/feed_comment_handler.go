package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tracker/internal/domain"
	"tracker/internal/middleware"
	"tracker/internal/repository"
)

type FeedCommentHandler struct {
	repo *repository.FeedCommentRepository
}

func NewFeedCommentHandler(repo *repository.FeedCommentRepository) *FeedCommentHandler {
	return &FeedCommentHandler{repo: repo}
}

// AddComment handles POST /feed/{eventId}/comments
func (h *FeedCommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		jsonError(w, "text is required", http.StatusBadRequest)
		return
	}

	fc := &domain.FeedComment{
		ID:          uuid.New(),
		FeedEventID: eventID,
		UserID:      userID,
		Text:        req.Text,
	}

	if err := h.repo.Create(r.Context(), fc); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, fc, http.StatusCreated)
}

// ListComments handles GET /feed/{eventId}/comments
func (h *FeedCommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	list, err := h.repo.ListForEvent(r.Context(), eventID)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []domain.FeedComment{}
	}
	jsonResponse(w, list, http.StatusOK)
}

// DeleteComment handles DELETE /feed/comments/{id}
func (h *FeedCommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), commentID, userID); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
