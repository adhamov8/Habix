package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tracker/internal/middleware"
	"tracker/internal/service"
)

type InteractionHandler struct {
	interactionSvc *service.InteractionService
}

func NewInteractionHandler(s *service.InteractionService) *InteractionHandler {
	return &InteractionHandler{interactionSvc: s}
}

func (h *InteractionHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	checkInID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid check-in id", http.StatusBadRequest)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		jsonError(w, "text is required", http.StatusBadRequest)
		return
	}

	comment, err := h.interactionSvc.AddComment(r.Context(), checkInID, userID, req.Text)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "check-in not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, comment, http.StatusCreated)
}

func (h *InteractionHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	checkInID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid check-in id", http.StatusBadRequest)
		return
	}

	comments, err := h.interactionSvc.GetComments(r.Context(), checkInID)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, comments, http.StatusOK)
}

func (h *InteractionHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	checkInID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid check-in id", http.StatusBadRequest)
		return
	}

	liked, err := h.interactionSvc.ToggleLike(r.Context(), checkInID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "check-in not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]bool{"liked": liked}, http.StatusOK)
}