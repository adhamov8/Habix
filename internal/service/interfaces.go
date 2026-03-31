package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"tracker/internal/domain"
	"tracker/internal/repository"
)

// Repository interfaces for testability

type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

type TokenRepo interface {
	Create(ctx context.Context, token *repository.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*repository.RefreshToken, error)
	Delete(ctx context.Context, hash string) error
}

type CheckInRepo interface {
	Create(ctx context.Context, ci *domain.SimpleCheckIn) error
	Delete(ctx context.Context, challengeID, userID uuid.UUID, date time.Time) error
	ExistsForDate(ctx context.Context, challengeID, userID uuid.UUID, date time.Time) (bool, error)
	ListForUser(ctx context.Context, challengeID, userID uuid.UUID) ([]domain.SimpleCheckIn, error)
	ListForChallenge(ctx context.Context, challengeID uuid.UUID) ([]domain.SimpleCheckIn, error)
}

type ChallengeRepo interface {
	Create(ctx context.Context, c *domain.Challenge) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Challenge, error)
	GetByInviteToken(ctx context.Context, token uuid.UUID) (*domain.Challenge, error)
	Update(ctx context.Context, c *domain.Challenge) error
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
	ListForUser(ctx context.Context, userID uuid.UUID) ([]domain.Challenge, error)
	ListPublic(ctx context.Context, categoryID *int, search string, limit, offset int) ([]domain.Challenge, error)
}

type ParticipantRepo interface {
	Add(ctx context.Context, challengeID, userID uuid.UUID) error
	Remove(ctx context.Context, challengeID, userID uuid.UUID) error
	Exists(ctx context.Context, challengeID, userID uuid.UUID) (bool, error)
	Count(ctx context.Context, challengeID uuid.UUID) (int, error)
	ListByChallenge(ctx context.Context, challengeID uuid.UUID) ([]domain.Participant, error)
}

type FeedRepo interface {
	Insert(ctx context.Context, e *domain.FeedEvent) error
	ListByChallenge(ctx context.Context, challengeID uuid.UUID, limit, offset int) ([]domain.FeedEvent, error)
}

type BadgeRepo interface {
	ListDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error)
	GetDefinitionByCode(ctx context.Context, code string) (*domain.BadgeDefinition, error)
	Award(ctx context.Context, userID uuid.UUID, badgeCode string, challengeID *uuid.UUID) (bool, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error)
	ListRecent(ctx context.Context, limit int) ([]domain.UserBadge, error)
	CountUserChallenges(ctx context.Context, userID uuid.UUID) (int, error)
}