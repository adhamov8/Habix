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
		WorkingDays: pq.Int64Array{0, 1, 2, 3, 4, 5, 6},
	}
}

func checkinSvc(ch *domain.Challenge, existsResult bool, participantExists bool) *CheckInService {
	cr := newMockChallengeRepo()
	cr.challenges[ch.ID] = ch
	return NewCheckInService(
		&mockCheckInRepo{existsResult: existsResult},
		cr,
		&mockParticipantRepoSimple{exists: participantExists},
		newMockFeedRepo(),
	)
}

// mockParticipantRepoSimple always returns a fixed exists value.
type mockParticipantRepoSimple struct {
	exists bool
}

func (m *mockParticipantRepoSimple) Add(_ context.Context, _, _ uuid.UUID) error    { return nil }
func (m *mockParticipantRepoSimple) Remove(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockParticipantRepoSimple) Exists(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return m.exists, nil
}
func (m *mockParticipantRepoSimple) Count(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }
func (m *mockParticipantRepoSimple) ListByChallenge(_ context.Context, _ uuid.UUID) ([]domain.Participant, error) {
	return nil, nil
}

// --- Tests ---

func TestCheckIn_Success(t *testing.T) {
	ch := activeChallenge()
	svc := checkinSvc(ch, false, true)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.NoError(t, err)
	assert.NotNil(t, ci)
	assert.Equal(t, ch.ID, ci.ChallengeID)
}

func TestCheckIn_AlreadyCheckedIn(t *testing.T) {
	ch := activeChallenge()
	svc := checkinSvc(ch, true, true)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrAlreadyChecked)
	assert.Nil(t, ci)
}

func TestCheckIn_NotWorkingDay(t *testing.T) {
	ch := activeChallenge()
	todayIdx := todayWorkingDay()
	otherDay := (todayIdx + 1) % 7
	ch.WorkingDays = pq.Int64Array{otherDay}

	svc := checkinSvc(ch, false, true)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrNotWorkingDay)
	assert.Nil(t, ci)
}

func TestCheckIn_NotParticipant(t *testing.T) {
	ch := activeChallenge()
	svc := checkinSvc(ch, false, false)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrNotParticipant)
	assert.Nil(t, ci)
}

func TestCheckIn_InactiveChallenge(t *testing.T) {
	ch := activeChallenge()
	ch.Status = "finished"

	svc := checkinSvc(ch, false, true)

	ci, err := svc.CheckIn(context.Background(), uuid.New(), ch.ID, "")
	assert.ErrorIs(t, err, ErrChallengeNotActive)
	assert.Nil(t, ci)
}