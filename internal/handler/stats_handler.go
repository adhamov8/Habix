package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tracker/internal/middleware"
	"tracker/internal/service"
)

type StatsHandler struct {
	statsSvc *service.StatsService
}

func NewStatsHandler(s *service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: s}
}

// Leaderboard godoc
// @Summary Get challenge leaderboard
// @Tags stats
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {array} domain.LeaderboardEntry
// @Router /challenges/{id}/leaderboard [get]
func (h *StatsHandler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	entries, err := h.statsSvc.Leaderboard(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "challenge not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		jsonResponse(w, []struct{}{}, http.StatusOK)
		return
	}
	jsonResponse(w, entries, http.StatusOK)
}

// ChallengeStats godoc
// @Summary Get challenge statistics
// @Tags stats
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {object} domain.ChallengeStats
// @Router /challenges/{id}/stats [get]
func (h *StatsHandler) ChallengeStats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	stats, err := h.statsSvc.ChallengeStats(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "challenge not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, stats, http.StatusOK)
}

// ChallengeSummary godoc
// @Summary Get finished challenge summary
// @Tags stats
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {object} domain.ChallengeSummary
// @Failure 400 {object} map[string]string
// @Router /challenges/{id}/summary [get]
func (h *StatsHandler) ChallengeSummary(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	summary, err := h.statsSvc.ChallengeSummary(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "challenge not found", http.StatusNotFound)
			return
		}
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, summary, http.StatusOK)
}

// PersonalStats godoc
// @Summary Get personal stats
// @Tags stats
// @Security BearerAuth
// @Success 200 {object} domain.PersonalStats
// @Router /users/me/stats [get]
func (h *StatsHandler) PersonalStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	stats, err := h.statsSvc.PersonalStats(r.Context(), userID)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, stats, http.StatusOK)
}