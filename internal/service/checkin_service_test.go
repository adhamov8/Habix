package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"tracker/internal/domain"
)

// --- Mock Repositories ---

type mockCheckInRepo struct {
	existsResult bool
	existsErr    error
	createErr    error
	created      []*domain.SimpleCheckIn
}

func (m *mockCheckInRepo) Create(_ context.Context, ci *domain.SimpleCheckIn) error {
	m.created = append(m.created, ci)
	return m.createErr
}
func (m *mockCheckInRepo) Delete(_ context.Context, _, _ uuid.UUID, _ time.Time) error { return nil }
func (m *mockCheckInRepo) ExistsForDate(_ context.Context, _, _ uuid.UUID, _ time.Time) (bool, error) {
	return m.existsResult, m.existsErr
}
func (m *mockCheckInRepo) ListForUser(_ context.Context, _, _ uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return nil, nil
}
func (m *mockCheckInRepo) ListForChallenge(_ context.Context, _ uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return nil, nil
}

type mockChallengeRepo struct {
	challenge *domain.Challenge
	getErr    error
	createErr error
}

func (m *mockChallengeRepo) Create(_ context.Context, c *domain.Challenge) error { return m.createErr }
func (m *mockChallengeRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Challenge, error) {
	if m.challenge == nil {
		return nil, sql.ErrNoRows
	}
	return m.challenge, m.getErr
}
func (m *mockChallengeRepo) GetByInviteToken(_ context.Context, _ uuid.UUID) (*domain.Challenge, error) {
	return m.challenge, m.getErr
}
func (m *mockChallengeRepo) Update(_ context.Context, _ *domain.Challenge) error { return nil }
func (m *mockChallengeRepo) SetStatus(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockChallengeRepo) ListForUser(_ context.Context, _ uuid.UUID) ([]domain.Challenge, error) {
	return nil, nil
}
func (m *mockChallengeRepo) ListPublic(_ context.Context, _ *int, _ string, _, _ int) ([]domain.Challenge, error) {
	return nil, nil
}

type mockParticipantRepo struct {
	exists    bool
	existsErr error
	addErr    error
}

func (m *mockParticipantRepo) Add(_ context.Context, _, _ uuid.UUID) error    { return m.addErr }
func (m *mockParticipantRepo) Remove(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockParticipantRepo) Exists(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return m.exists, m.existsErr
}
func (m *mockParticipantRepo) Count(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockParticipantRepo) ListByChallenge(_ context.Context, _ uuid.UUID) ([]domain.Participant, error) {
	return nil, nil
}

type mockFeedRepo struct {
	events []*domain.FeedEvent
}

func (m *mockFeedRepo) Insert(_ context.Context, e *domain.FeedEvent) error {
	m.events = append(m.events, e)
	return nil
}
func (m *mockFeedRepo) ListByChallenge(_ context.Context, _ uuid.UUID, _, _ int) ([]domain.FeedEvent, error) {
	return nil, nil
}

// --- Helper ---

func todayWorkingDay() int64 {
	return int64((int(time.Now().UTC().Weekday()) + 6) % 7)
}

func activeChallenge() *domain.Challenge {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	return &domain.Challenge{
		ID:          uuid.New(),
		CreatorID:   uuid.New(),
		Status:      "active",
		StartsAt:    today.AddDate(0, 0, -10),
		EndsAt:      today.AddDate(0, 0, 10),
		WorkingDays: pq.Int64Array{0, 1, 2, 3, 4, 5, 6}, // every day
	}
}

// --- Tests ---

func TestCheckIn_Success(t *testing.T) {
	ch := activeChallenge()
	svc := NewCheckInService(
		&mockCheckInRepo{existsResult: false},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true},
		&mockFeedRepo{},
	)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.NoError(t, err)
	assert.NotNil(t, ci)
	assert.Equal(t, ch.ID, ci.ChallengeID)
}

func TestCheckIn_AlreadyCheckedIn(t *testing.T) {
	ch := activeChallenge()
	svc := NewCheckInService(
		&mockCheckInRepo{existsResult: true},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true},
		&mockFeedRepo{},
	)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrAlreadyChecked)
	assert.Nil(t, ci)
}

func TestCheckIn_NotWorkingDay(t *testing.T) {
	ch := activeChallenge()
	// Set working days to NOT include today
	todayIdx := todayWorkingDay()
	otherDay := (todayIdx + 1) % 7
	ch.WorkingDays = pq.Int64Array{otherDay}

	svc := NewCheckInService(
		&mockCheckInRepo{existsResult: false},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true},
		&mockFeedRepo{},
	)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrNotWorkingDay)
	assert.Nil(t, ci)
}

func TestCheckIn_NotParticipant(t *testing.T) {
	ch := activeChallenge()
	svc := NewCheckInService(
		&mockCheckInRepo{existsResult: false},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: false},
		&mockFeedRepo{},
	)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrNotParticipant)
	assert.Nil(t, ci)
}

func TestCheckIn_InactiveChallenge(t *testing.T) {
	ch := activeChallenge()
	ch.Status = "finished"

	svc := NewCheckInService(
		&mockCheckInRepo{existsResult: false},
		&mockChallengeRepo{challenge: ch},
		&mockParticipantRepo{exists: true},
		&mockFeedRepo{},
	)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrChallengeNotActive)
	assert.Nil(t, ci)
}