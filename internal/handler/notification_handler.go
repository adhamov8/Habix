package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tracker/internal/middleware"
	"tracker/internal/service"
)

type NotificationHandler struct {
	notifSvc *service.NotificationService
}

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: s}
}

// List handles GET /notifications
// @Summary List notifications
// @Tags notifications
// @Security BearerAuth
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Notification
// @Router /notifications [get]
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	list, err := h.notifSvc.List(r.Context(), userID, limit, offset)
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

// UnreadCount handles GET /notifications/unread-count
// @Summary Get unread notification count
// @Tags notifications
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	count, err := h.notifSvc.CountUnread(r.Context(), userID)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]int{"count": count}, http.StatusOK)
}

// MarkRead handles PATCH /notifications/{id}/read
// @Summary Mark notification as read
// @Tags notifications
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 204
// @Router /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	if err := h.notifSvc.MarkRead(r.Context(), id, userID); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MarkAllRead handles PATCH /notifications/read-all
// @Summary Mark all notifications as read
// @Tags notifications
// @Security BearerAuth
// @Success 204
// @Router /notifications/read-all [patch]
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	if err := h.notifSvc.MarkAllRead(r.Context(), userID); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}