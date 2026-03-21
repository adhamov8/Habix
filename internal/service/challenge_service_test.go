package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreate_Success(t *testing.T) {
	svc := NewChallengeService(
		&mockChallengeRepo{},
		&mockParticipantRepo{},
		&mockFeedRepo{},
	)

	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02")
	futureEnd := time.Now().UTC().AddDate(0, 1, 0).Format("2006-01-02")

	ch, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID:  1,
		Title:       "Тест",
		StartsAt:    future,
		EndsAt:      futureEnd,
		WorkingDays: []int64{0, 1, 2, 3, 4},
	})

	assert.NoError(t, err)
	assert.NotNil(t, ch)
	assert.Equal(t, "Тест", ch.Title)
	assert.Equal(t, "upcoming", ch.Status)
}

func TestCreate_InvalidDates(t *testing.T) {
	svc := NewChallengeService(
		&mockChallengeRepo{},
		&mockParticipantRepo{},
		&mockFeedRepo{},
	)

	_, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Тест",
		StartsAt:   "not-a-date",
		EndsAt:     "2025-12-31",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid starts_at")
}

func TestJoinPublic_Success(t *testing.T) {
	challengeID := uuid.New()
	ch := activeChallenge()
	ch.ID = challengeID
	ch.IsPublic = true

	svc := NewChallengeService(
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: false},
		&mockFeedRepo{},
	)

	err := svc.JoinPublic(context.Background(), uuid.New(), challengeID)
	assert.NoError(t, err)
}

func TestJoinPublic_AlreadyJoined(t *testing.T) {
	challengeID := uuid.New()
	ch := activeChallenge()
	ch.ID = challengeID
	ch.IsPublic = true

	svc := NewChallengeService(
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true}, // already joined
		&mockFeedRepo{},
	)

	err := svc.JoinPublic(context.Background(), uuid.New(), challengeID)
	assert.ErrorIs(t, err, ErrAlreadyJoined)
}