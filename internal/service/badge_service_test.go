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

// --- Mock Badge Repo ---

type mockBadgeRepo struct {
	awarded     map[string]bool // code -> awarded
	definitions []domain.BadgeDefinition
	userBadges  []domain.UserBadge
	challenges  int
}

func newMockBadgeRepo() *mockBadgeRepo {
	return &mockBadgeRepo{
		awarded: make(map[string]bool),
		definitions: []domain.BadgeDefinition{
			{ID: 1, Code: "first_checkin", Title: "Первый шаг", Icon: "✅"},
			{ID: 2, Code: "streak_3", Title: "Три дня подряд", Icon: "🔥"},
			{ID: 3, Code: "streak_7", Title: "Неделя огня", Icon: "⚡"},
			{ID: 4, Code: "streak_30", Title: "Месяц дисциплины", Icon: "💎"},
		},
	}
}

func (m *mockBadgeRepo) ListDefinitions(_ context.Context) ([]domain.BadgeDefinition, error) {
	return m.definitions, nil
}
func (m *mockBadgeRepo) GetDefinitionByCode(_ context.Context, code string) (*domain.BadgeDefinition, error) {
	for _, bd := range m.definitions {
		if bd.Code == code {
			return &bd, nil
		}
	}
	return nil, nil
}
func (m *mockBadgeRepo) Award(_ context.Context, _ uuid.UUID, code string, _ *uuid.UUID) (bool, error) {
	if m.awarded[code] {
		return false, nil // duplicate
	}
	m.awarded[code] = true
	return true, nil
}
func (m *mockBadgeRepo) ListForUser(_ context.Context, _ uuid.UUID) ([]domain.UserBadge, error) {
	return m.userBadges, nil
}
func (m *mockBadgeRepo) ListRecent(_ context.Context, _ int) ([]domain.UserBadge, error) {
	return nil, nil
}
func (m *mockBadgeRepo) CountUserChallenges(_ context.Context, _ uuid.UUID) (int, error) {
	return m.challenges, nil
}

// --- Mock CheckIn Repo for badge tests ---

type mockBadgeCheckInRepo struct {
	checkIns []domain.SimpleCheckIn
}

func (m *mockBadgeCheckInRepo) Create(_ context.Context, _ *domain.SimpleCheckIn) error { return nil }
func (m *mockBadgeCheckInRepo) Delete(_ context.Context, _, _ uuid.UUID, _ time.Time) error {
	return nil
}
func (m *mockBadgeCheckInRepo) ExistsForDate(_ context.Context, _, _ uuid.UUID, _ time.Time) (bool, error) {
	return false, nil
}
func (m *mockBadgeCheckInRepo) ListForUser(_ context.Context, _, _ uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return m.checkIns, nil
}
func (m *mockBadgeCheckInRepo) ListForChallenge(_ context.Context, _ uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return nil, nil
}

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

	badgeRepo := newMockBadgeRepo()
	svc := NewBadgeService(
		badgeRepo,
		&mockBadgeCheckInRepo{
			checkIns: []domain.SimpleCheckIn{
				{ID: uuid.New(), ChallengeID: challengeID, UserID: userID, Date: today},
			},
		},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true},
		&mockFeedRepo{},
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

	// Create 7 consecutive check-ins ending today
	var checkIns []domain.SimpleCheckIn
	for i := 6; i >= 0; i-- {
		checkIns = append(checkIns, domain.SimpleCheckIn{
			ID:          uuid.New(),
			ChallengeID: challengeID,
			UserID:      userID,
			Date:        today.AddDate(0, 0, -i),
		})
	}

	badgeRepo := newMockBadgeRepo()
	svc := NewBadgeService(
		badgeRepo,
		&mockBadgeCheckInRepo{checkIns: checkIns},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true},
		&mockFeedRepo{},
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

	badgeRepo := newMockBadgeRepo()
	// Pre-award badge to simulate duplicate
	badgeRepo.awarded["first_checkin"] = true

	svc := NewBadgeService(
		badgeRepo,
		&mockBadgeCheckInRepo{
			checkIns: []domain.SimpleCheckIn{
				{ID: uuid.New(), ChallengeID: challengeID, UserID: userID, Date: today},
			},
		},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true},
		&mockFeedRepo{},
	)

	feedRepo := svc.feed.(*mockFeedRepo)
	svc.CheckAndAward(context.Background(), userID, challengeID)

	// Feed event should NOT be created for duplicate badge
	hasBadgeFeed := false
	for _, e := range feedRepo.events {
		if e.Type == "badge_earned" {
			hasBadgeFeed = true
		}
	}
	assert.False(t, hasBadgeFeed, "should not create feed event for duplicate badge")
}