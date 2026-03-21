package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tracker/internal/middleware"
	"tracker/internal/service"
)

type BadgeHandler struct {
	badgeSvc *service.BadgeService
}

func NewBadgeHandler(s *service.BadgeService) *BadgeHandler {
	return &BadgeHandler{badgeSvc: s}
}

// ListDefinitions handles GET /badges
// @Summary List all badge definitions
// @Tags badges
// @Security BearerAuth
// @Success 200 {array} domain.BadgeDefinition
// @Router /badges [get]
func (h *BadgeHandler) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	list, err := h.badgeSvc.ListDefinitions(r.Context())
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

// MyBadges handles GET /users/me/badges
// @Summary Get current user's badges
// @Tags badges
// @Security BearerAuth
// @Success 200 {array} domain.UserBadge
// @Router /users/me/badges [get]
func (h *BadgeHandler) MyBadges(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	list, err := h.badgeSvc.GetUserBadges(r.Context(), userID)
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

// UserBadges handles GET /users/{id}/badges
// @Summary Get user's badges
// @Tags badges
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {array} domain.UserBadge
// @Router /users/{id}/badges [get]
func (h *BadgeHandler) UserBadges(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}
	list, err := h.badgeSvc.GetUserBadges(r.Context(), id)
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