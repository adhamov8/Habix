package service

import (
	"context"

	"tracker/internal/domain"

	"github.com/google/uuid"
)

type FeedService struct {
	feed         FeedRepo
	participants ParticipantRepo
}

func NewFeedService(f FeedRepo, p ParticipantRepo) *FeedService {
	return &FeedService{feed: f, participants: p}
}

func (s *FeedService) GetFeed(ctx context.Context, challengeID, userID uuid.UUID, limit, offset int) ([]domain.FeedEvent, error) {
	// Только участники и автор могут читать ленту
	isParticipant, err := s.participants.Exists(ctx, challengeID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrForbidden
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.feed.ListByChallenge(ctx, challengeID, limit, offset)
}

// помощник для добавления событий в ленту из других сервисов и хендлеров
func (s *FeedService) InsertEvent(ctx context.Context, challengeID, userID uuid.UUID, eventType string, refID *uuid.UUID) {
	_ = s.feed.Insert(ctx, &domain.FeedEvent{
		ID:          uuid.New(),
		ChallengeID: challengeID,
		UserID:      userID,
		Type:        eventType,
		ReferenceID: refID,
	})
}
