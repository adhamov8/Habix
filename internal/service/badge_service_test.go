package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"tracker/internal/domain"
)

// --- Tests ---

func TestCheckAndAward_FirstCheckin(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	challengeID := uuid.New()
	userID := uuid.New()

	ch := &domain.Challenge{
		ID:          challengeID,
		Status:      "active",
		StartsAt:    today.AddDate(0, 0, -5),
		EndsAt:      today.AddDate(0, 0, 25),
		WorkingDays: pq.Int64Array{0, 1, 2, 3, 4, 5, 6},
	}

	cr := newMockChallengeRepo()
	cr.challenges[ch.ID] = ch
	pr := newMockParticipantRepo()
	pr.participants[pr.key(ch.ID, userID)] = true

	badgeRepo := newMockBadgeRepo()
	svc := NewBadgeService(
		badgeRepo,
		&mockCheckInRepo{
			checkIns: []domain.SimpleCheckIn{
				{ID: uuid.New(), ChallengeID: challengeID, UserID: userID, Date: today},
			},
		},
		cr,
		pr,
		newMockFeedRepo(),
	)

	svc.CheckAndAward(context.Background(), userID, challengeID)

	assert.True(t, badgeRepo.awarded["first_checkin"], "should award first_checkin badge")
}

func TestCheckAndAward_Streak7(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	challengeID := uuid.New()
	userID := uuid.New()

	ch := &domain.Challenge{
		ID:          challengeID,
		Status:      "active",
		StartsAt:    today.AddDate(0, 0, -10),
		EndsAt:      today.AddDate(0, 0, 20),
		WorkingDays: pq.Int64Array{0, 1, 2, 3, 4, 5, 6},
	}

	var checkIns []domain.SimpleCheckIn
	for i := 6; i >= 0; i-- {
		checkIns = append(checkIns, domain.SimpleCheckIn{
			ID:          uuid.New(),
			ChallengeID: challengeID,
			UserID:      userID,
			Date:        today.AddDate(0, 0, -i),
		})
	}

	cr := newMockChallengeRepo()
	cr.challenges[ch.ID] = ch
	pr := newMockParticipantRepo()
	pr.participants[pr.key(ch.ID, userID)] = true

	badgeRepo := newMockBadgeRepo()
	svc := NewBadgeService(
		badgeRepo,
		&mockCheckInRepo{checkIns: checkIns},
		cr,
		pr,
		newMockFeedRepo(),
	)

	svc.CheckAndAward(context.Background(), userID, challengeID)

	assert.True(t, badgeRepo.awarded["streak_3"], "should award streak_3")
	assert.True(t, badgeRepo.awarded["streak_7"], "should award streak_7")
	assert.False(t, badgeRepo.awarded["streak_30"], "should NOT award streak_30")
}

func TestCheckAndAward_NoDuplicate(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	challengeID := uuid.New()
	userID := uuid.New()

	ch := &domain.Challenge{
		ID:          challengeID,
		Status:      "active",
		StartsAt:    today.AddDate(0, 0, -5),
		EndsAt:      today.AddDate(0, 0, 25),
		WorkingDays: pq.Int64Array{0, 1, 2, 3, 4, 5, 6},
	}

	cr := newMockChallengeRepo()
	cr.challenges[ch.ID] = ch
	pr := newMockParticipantRepo()
	pr.participants[pr.key(ch.ID, userID)] = true

	badgeRepo := newMockBadgeRepo()
	badgeRepo.awarded["first_checkin"] = true

	fr := newMockFeedRepo()
	svc := NewBadgeService(
		badgeRepo,
		&mockCheckInRepo{
			checkIns: []domain.SimpleCheckIn{
				{ID: uuid.New(), ChallengeID: challengeID, UserID: userID, Date: today},
			},
		},
		cr,
		pr,
		fr,
	)

	svc.CheckAndAward(context.Background(), userID, challengeID)

	hasBadgeFeed := false
	for _, e := range fr.events {
		if e.Type == "badge_earned" {
			hasBadgeFeed = true
		}
	}
	assert.False(t, hasBadgeFeed, "should not create feed event for duplicate badge")
}