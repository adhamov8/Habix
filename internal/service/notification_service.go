package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"tracker/internal/domain"
	"tracker/internal/repository"
)

type NotificationService struct {
	notifications *repository.NotificationRepository
}

func NewNotificationService(n *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: n}
}

func (s *NotificationService) Notify(ctx context.Context, userID uuid.UUID, nType, title, body string, data map[string]interface{}) error {
	n := &domain.Notification{
		UserID: userID,
		Type:   nType,
		Title:  title,
		Body:   body,
	}
	if data != nil {
		raw, err := json.Marshal(data)
		if err == nil {
			msg := json.RawMessage(raw)
			n.Data = &msg
		}
	}
	return s.notifications.Create(ctx, n)
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Notification, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.notifications.ListForUser(ctx, userID, limit, offset)
}

func (s *NotificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.notifications.CountUnread(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.notifications.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifications.MarkAllRead(ctx, userID)
}