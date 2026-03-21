package service

import (
	"context"

	"github.com/google/uuid"
	"tracker/internal/domain"
	"tracker/internal/repository"
)

type InteractionService struct {
	comments *repository.CommentRepository
	likes    *repository.LikeRepository
}

func NewInteractionService(c *repository.CommentRepository, l *repository.LikeRepository) *InteractionService {
	return &InteractionService{comments: c, likes: l}
}

func (s *InteractionService) AddComment(ctx context.Context, checkInID, userID uuid.UUID, text string) (*domain.Comment, error) {
	c := &domain.Comment{
		ID:        uuid.New(),
		CheckInID: checkInID,
		UserID:    userID,
		Text:      text,
	}
	if err := s.comments.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *InteractionService) GetComments(ctx context.Context, checkInID uuid.UUID) ([]domain.Comment, error) {
	return s.comments.ListByCheckIn(ctx, checkInID)
}

func (s *InteractionService) ToggleLike(ctx context.Context, checkInID, userID uuid.UUID) (bool, error) {
	return s.likes.Toggle(ctx, checkInID, userID)
}