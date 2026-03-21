package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"tracker/internal/domain"
	"tracker/internal/middleware"
	"tracker/internal/repository"
	"tracker/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ChallengeHandler struct {
	challengeSvc *service.ChallengeService
	categories   *repository.CategoryRepository
}

func NewChallengeHandler(cs *service.ChallengeService, cats *repository.CategoryRepository) *ChallengeHandler {
	return &ChallengeHandler{challengeSvc: cs, categories: cats}
}

// ListCategories godoc
// @Summary List all categories
// @Tags categories
// @Security BearerAuth
// @Success 200 {array} domain.Category
// @Router /categories [get]
func (h *ChallengeHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categories.List(r.Context())
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if cats == nil {
		jsonResponse(w, []struct{}{}, http.StatusOK)
		return
	}
	jsonResponse(w, cats, http.StatusOK)
}

// Create godoc
// @Summary Create a new challenge
// @Tags challenges
// @Security BearerAuth
// @Param body body service.CreateChallengeParams true "Challenge data"
// @Success 201 {object} domain.Challenge
// @Failure 400 {object} map[string]string
// @Router /challenges [post]
func (h *ChallengeHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var p service.CreateChallengeParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if p.Title == "" || p.StartsAt == "" || p.EndsAt == "" || p.CategoryID == 0 {
		jsonError(w, "title, category_id, starts_at and ends_at are required", http.StatusBadRequest)
		return
	}

	c, err := h.challengeSvc.Create(r.Context(), userID, p)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, c, http.StatusCreated)
}

// ListChallenges godoc
// @Summary List public challenges
// @Tags challenges
// @Security BearerAuth
// @Param public query string false "Filter public"
// @Param category query int false "Category ID"
// @Param search query string false "Search term"
// @Param page query int false "Page number"
// @Success 200 {array} domain.Challenge
// @Router /challenges [get]
func (h *ChallengeHandler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if q.Get("public") == "true" {
		var categoryID *int
		if cat := q.Get("category"); cat != "" {
			v, err := strconv.Atoi(cat)
			if err != nil {
				jsonError(w, "invalid category param", http.StatusBadRequest)
				return
			}
			categoryID = &v
		}
		search := q.Get("search")
		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}
		limit := 20
		offset := (page - 1) * limit

		list, err := h.challengeSvc.ListPublic(r.Context(), categoryID, search, limit, offset)
		if err != nil {
			jsonError(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if list == nil {
			jsonResponse(w, []struct{}{}, http.StatusOK)
			return
		}
		jsonResponse(w, list, http.StatusOK)
		return
	}

	// Default: list all public challenges (no filter)
	list, err := h.challengeSvc.ListPublic(r.Context(), nil, "", 20, 0)
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

// ListMy godoc
// @Summary List my challenges
// @Tags challenges
// @Security BearerAuth
// @Success 200 {array} domain.Challenge
// @Router /challenges/my [get]
func (h *ChallengeHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	list, err := h.challengeSvc.ListForUser(r.Context(), userID)
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

// GetByID godoc
// @Summary Get challenge by ID
// @Tags challenges
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {object} domain.Challenge
// @Failure 404 {object} map[string]string
// @Router /challenges/{id} [get]
func (h *ChallengeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	c, err := h.challengeSvc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "challenge not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Compute status on the fly, but never overwrite manual "finished"
	if c.Status != "finished" {
		now := time.Now().UTC().Truncate(24 * time.Hour)
		start := c.StartsAt.Truncate(24 * time.Hour)
		end := c.EndsAt.Truncate(24 * time.Hour)
		computed := "active"
		if now.Before(start) {
			computed = "upcoming"
		} else if now.After(end) {
			computed = "finished"
		}
		if c.Status != computed {
			c.Status = computed
			_ = h.challengeSvc.UpdateStatus(r.Context(), c.ID, computed)
		}
	}

	isParticipant := h.challengeSvc.IsParticipant(r.Context(), id, userID)
	isCreator := c.CreatorID == userID
	participantCount := h.challengeSvc.ParticipantCount(r.Context(), id)

	type resp struct {
		*domain.Challenge
		IsParticipant    bool `json:"is_participant"`
		IsCreator        bool `json:"is_creator"`
		ParticipantCount int  `json:"participant_count"`
	}
	jsonResponse(w, resp{Challenge: c, IsParticipant: isParticipant, IsCreator: isCreator, ParticipantCount: participantCount}, http.StatusOK)
}

// Update godoc
// @Summary Update a challenge
// @Tags challenges
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Param body body service.UpdateChallengeParams true "Update data"
// @Success 200 {object} domain.Challenge
// @Router /challenges/{id} [patch]
func (h *ChallengeHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}

	var p service.UpdateChallengeParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, err := h.challengeSvc.Update(r.Context(), id, userID, p)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	jsonResponse(w, c, http.StatusOK)
}

// Finish godoc
// @Summary Finish a challenge
// @Tags challenges
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 204
// @Router /challenges/{id}/finish [post]
func (h *ChallengeHandler) Finish(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	if err := h.challengeSvc.Finish(r.Context(), id, userID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetInviteLink godoc
// @Summary Get invite link for a challenge
// @Tags challenges
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {object} map[string]string
// @Router /challenges/{id}/invite-link [get]
func (h *ChallengeHandler) GetInviteLink(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	token, err := h.challengeSvc.GetInviteLink(r.Context(), id, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	jsonResponse(w, map[string]string{"invite_token": token}, http.StatusOK)
}

// JoinByInvite godoc
// @Summary Join challenge by invite token
// @Tags challenges
// @Security BearerAuth
// @Param inviteToken path string true "Invite token"
// @Success 204
// @Router /challenges/join/{inviteToken} [post]
func (h *ChallengeHandler) JoinByInvite(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	token, err := uuid.Parse(chi.URLParam(r, "inviteToken"))
	if err != nil {
		jsonError(w, "invalid invite token", http.StatusBadRequest)
		return
	}
	c, err := h.challengeSvc.JoinByInviteToken(r.Context(), userID, token)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyJoined) && c != nil {
			jsonResponse(w, map[string]string{"challenge_id": c.ID.String()}, http.StatusConflict)
			return
		}
		handleServiceError(w, err)
		return
	}
	jsonResponse(w, map[string]string{"challenge_id": c.ID.String()}, http.StatusOK)
}

// JoinPublic godoc
// @Summary Join a public challenge
// @Tags challenges
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 204
// @Router /challenges/{id}/join [post]
func (h *ChallengeHandler) JoinPublic(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	if err := h.challengeSvc.JoinPublic(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChallengeHandler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}
	targetID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}
	if err := h.challengeSvc.RemoveParticipant(r.Context(), challengeID, userID, targetID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		jsonError(w, "not found", http.StatusNotFound)
	case errors.Is(err, service.ErrForbidden):
		jsonError(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, service.ErrNotUpcoming):
		jsonError(w, err.Error(), http.StatusConflict)
	case errors.Is(err, service.ErrAlreadyJoined):
		jsonError(w, err.Error(), http.StatusConflict)
	case errors.Is(err, service.ErrNotPublic):
		jsonError(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, service.ErrChallengeEnded):
		jsonError(w, err.Error(), http.StatusGone)
	default:
		jsonError(w, err.Error(), http.StatusBadRequest)
	}
}
