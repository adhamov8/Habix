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

type CheckInHandler struct {
	checkInSvc *service.CheckInService
}

func NewCheckInHandler(s *service.CheckInService) *CheckInHandler {
	return &CheckInHandler{checkInSvc: s}
}

// CheckIn handles POST /challenges/{id}/checkin
// @Summary Check in today
// @Tags checkin
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 201 {object} domain.SimpleCheckIn
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /challenges/{id}/checkin [post]
func (h *CheckInHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	// Body is optional — empty body means empty comment
	_ = json.NewDecoder(r.Body).Decode(&req)

	ci, err := h.checkInSvc.CheckIn(r.Context(), userID, challengeID, req.Comment)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			jsonError(w, "challenge not found", http.StatusNotFound)
		case errors.Is(err, service.ErrChallengeNotActive):
			jsonError(w, err.Error(), http.StatusConflict)
		case errors.Is(err, service.ErrNotParticipant):
			jsonError(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, service.ErrNotWorkingDay):
			jsonError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrAlreadyChecked):
			jsonError(w, err.Error(), http.StatusConflict)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, ci, http.StatusCreated)
}

// Undo handles DELETE /challenges/{id}/checkin
// @Summary Undo today's check-in
// @Tags checkin
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 204
// @Failure 409 {object} map[string]string
// @Router /challenges/{id}/checkin [delete]
func (h *CheckInHandler) Undo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}

	if err := h.checkInSvc.Undo(r.Context(), userID, challengeID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotCheckedIn):
			jsonError(w, err.Error(), http.StatusConflict)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetProgress handles GET /challenges/{id}/progress
// @Summary Get user progress in a challenge
// @Tags checkin
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {object} domain.Progress
// @Failure 404 {object} map[string]string
// @Router /challenges/{id}/progress [get]
func (h *CheckInHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}

	progress, err := h.checkInSvc.GetProgress(r.Context(), userID, challengeID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "challenge not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, progress, http.StatusOK)
}

// ListAll handles GET /challenges/{id}/checkins
// @Summary List all user check-ins in a challenge
// @Tags checkin
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {array} domain.SimpleCheckIn
// @Router /challenges/{id}/checkins [get]
func (h *CheckInHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}

	list, err := h.checkInSvc.ListAll(r.Context(), challengeID, userID)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		jsonResponse(w, []struct{}{}, http.StatusOK)
		return
	}
	jsonResponse(w, list, http.StatusOK)
}