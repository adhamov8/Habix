package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tracker/internal/middleware"
	"tracker/internal/service"
)

type FeedHandler struct {
	feedSvc *service.FeedService
}

func NewFeedHandler(s *service.FeedService) *FeedHandler {
	return &FeedHandler{feedSvc: s}
}

func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	events, err := h.feedSvc.GetFeed(r.Context(), challengeID, userID, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			jsonError(w, "forbidden", http.StatusForbidden)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, events, http.StatusOK)
}