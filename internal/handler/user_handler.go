package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tracker/internal/middleware"
	"tracker/internal/repository"
	"tracker/internal/service"
)

type UserHandler struct {
	users    *repository.UserRepository
	statsSvc *service.StatsService
}

func NewUserHandler(users *repository.UserRepository, statsSvc *service.StatsService) *UserHandler {
	return &UserHandler{users: users, statsSvc: statsSvc}
}

// GetMe godoc
// @Summary Get current user
// @Tags users
// @Security BearerAuth
// @Success 200 {object} domain.User
// @Router /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, user, http.StatusOK)
}

// UpdateMe godoc
// @Summary Update current user profile
// @Tags users
// @Security BearerAuth
// @Param body body object true "Update data"
// @Success 200 {object} domain.User
// @Router /users/me [patch]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var req struct {
		Name     *string `json:"name"`
		Bio      *string `json:"bio"`
		Timezone *string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	if err := h.users.Update(r.Context(), user); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, user, http.StatusOK)
}

// GetProfile godoc
// @Summary Get user public profile
// @Tags users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} object
// @Router /users/{id}/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	stats, err := h.statsSvc.PersonalStats(r.Context(), id)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type profileResponse struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Bio       *string   `json:"bio"`
		CreatedAt string    `json:"created_at"`
		Stats     any       `json:"stats"`
	}

	jsonResponse(w, profileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt.Format("2006-01-02"),
		Stats:     stats,
	}, http.StatusOK)
}