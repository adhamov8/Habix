package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func newTestChallengeService() (*ChallengeService, *mockChallengeRepo, *mockParticipantRepo, *mockFeedRepo) {
	cr := newMockChallengeRepo()
	pr := newMockParticipantRepo()
	fr := newMockFeedRepo()
	return NewChallengeService(cr, pr, fr), cr, pr, fr
}

func futureDate(days int) string {
	return time.Now().UTC().AddDate(0, 0, days).Format("2006-01-02")
}

// --- Create tests ---

func TestCreate_Success(t *testing.T) {
	svc, cr, pr, fr := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Morning Run",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)
	assert.Equal(t, "Morning Run", challenge.Title)
	assert.Equal(t, creatorID, challenge.CreatorID)
	assert.Equal(t, "upcoming", challenge.Status)
	assert.Equal(t, "23:00", challenge.DeadlineTime)

	assert.Len(t, cr.challenges, 1)

	joined, _ := pr.Exists(context.Background(), challenge.ID, creatorID)
	assert.True(t, joined)

	assert.Len(t, fr.events, 1)
	assert.Equal(t, "challenge_created", fr.events[0].Type)
}

func TestCreate_ActiveWhenStartsToday(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Today Start",
		StartsAt:   time.Now().UTC().Format("2006-01-02"),
		EndsAt:     futureDate(10),
	})
	require.NoError(t, err)
	assert.Equal(t, "active", challenge.Status)
}

func TestCreate_InvalidDateFormat(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	_, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Bad Date",
		StartsAt:   "not-a-date",
		EndsAt:     futureDate(10),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid starts_at")
}

// --- Update tests ---

func TestUpdate_Success(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Original",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	newTitle := "Updated Title"
	updated, err := svc.Update(context.Background(), challenge.ID, creatorID, UpdateChallengeParams{
		Title: &newTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
}

func TestUpdate_Forbidden(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Test",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	newTitle := "Hacked"
	_, err = svc.Update(context.Background(), challenge.ID, uuid.New(), UpdateChallengeParams{
		Title: &newTitle,
	})
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUpdate_NotUpcoming(t *testing.T) {
	svc, cr, _, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Test",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	cr.challenges[challenge.ID].Status = "active"

	newTitle := "Update"
	_, err = svc.Update(context.Background(), challenge.ID, creatorID, UpdateChallengeParams{
		Title: &newTitle,
	})
	assert.ErrorIs(t, err, ErrNotUpcoming)
}

func TestUpdate_NotFound(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	newTitle := "X"
	_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), UpdateChallengeParams{
		Title: &newTitle,
	})
	assert.ErrorIs(t, err, ErrNotFound)
}

// --- Finish tests ---

func TestFinish_Success(t *testing.T) {
	svc, cr, _, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Finishable",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	err = svc.Finish(context.Background(), challenge.ID, creatorID)
	require.NoError(t, err)
	assert.Equal(t, "finished", cr.challenges[challenge.ID].Status)
}

func TestFinish_Forbidden(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Test",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	err = svc.Finish(context.Background(), challenge.ID, uuid.New())
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestFinish_NotFound(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()
	err := svc.Finish(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

// --- GetByID tests ---

func TestGetByID_Success(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	created, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Lookup",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	found, err := svc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestGetByID_NotFound(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()
	_, err := svc.GetByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

// --- JoinPublic tests ---

func TestJoinPublic_Success(t *testing.T) {
	svc, _, pr, fr := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Public",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
		IsPublic:   true,
	})
	require.NoError(t, err)

	joiner := uuid.New()
	err = svc.JoinPublic(context.Background(), joiner, challenge.ID)
	require.NoError(t, err)

	joined, _ := pr.Exists(context.Background(), challenge.ID, joiner)
	assert.True(t, joined)

	assert.Len(t, fr.events, 2)
	assert.Equal(t, "user_joined", fr.events[1].Type)
}

func TestJoinPublic_NotPublic(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Private",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
		IsPublic:   false,
	})
	require.NoError(t, err)

	err = svc.JoinPublic(context.Background(), uuid.New(), challenge.ID)
	assert.ErrorIs(t, err, ErrNotPublic)
}

func TestJoinPublic_AlreadyJoined(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Public",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
		IsPublic:   true,
	})
	require.NoError(t, err)

	err = svc.JoinPublic(context.Background(), creatorID, challenge.ID)
	assert.ErrorIs(t, err, ErrAlreadyJoined)
}

func TestJoinPublic_Finished(t *testing.T) {
	svc, cr, _, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Done",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
		IsPublic:   true,
	})
	require.NoError(t, err)

	cr.challenges[challenge.ID].Status = "finished"

	err = svc.JoinPublic(context.Background(), uuid.New(), challenge.ID)
	assert.ErrorIs(t, err, ErrChallengeEnded)
}

// --- JoinByInviteToken tests ---

func TestJoinByInviteToken_Success(t *testing.T) {
	svc, _, pr, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Invite",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	joiner := uuid.New()
	result, err := svc.JoinByInviteToken(context.Background(), joiner, challenge.InviteToken)
	require.NoError(t, err)
	assert.Equal(t, challenge.ID, result.ID)

	joined, _ := pr.Exists(context.Background(), challenge.ID, joiner)
	assert.True(t, joined)
}

func TestJoinByInviteToken_InvalidToken(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	_, err := svc.JoinByInviteToken(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

// --- RemoveParticipant tests ---

func TestRemoveParticipant_Success(t *testing.T) {
	svc, _, pr, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Remove Test",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	target := uuid.New()
	_ = pr.Add(context.Background(), challenge.ID, target)

	err = svc.RemoveParticipant(context.Background(), challenge.ID, creatorID, target)
	require.NoError(t, err)

	exists, _ := pr.Exists(context.Background(), challenge.ID, target)
	assert.False(t, exists)
}

func TestRemoveParticipant_CannotRemoveSelf(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Self Remove",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	err = svc.RemoveParticipant(context.Background(), challenge.ID, creatorID, creatorID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creator cannot remove themselves")
}

func TestRemoveParticipant_Forbidden(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Test",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	err = svc.RemoveParticipant(context.Background(), challenge.ID, uuid.New(), uuid.New())
	assert.ErrorIs(t, err, ErrForbidden)
}

// --- GetInviteLink tests ---

func TestGetInviteLink_Creator(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Invite Link",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	link, err := svc.GetInviteLink(context.Background(), challenge.ID, creatorID)
	require.NoError(t, err)
	assert.Equal(t, challenge.InviteToken.String(), link)
}

func TestGetInviteLink_Participant(t *testing.T) {
	svc, _, pr, _ := newTestChallengeService()
	creatorID := uuid.New()

	challenge, err := svc.Create(context.Background(), creatorID, CreateChallengeParams{
		CategoryID: 1,
		Title:      "Invite Link",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	participant := uuid.New()
	_ = pr.Add(context.Background(), challenge.ID, participant)

	link, err := svc.GetInviteLink(context.Background(), challenge.ID, participant)
	require.NoError(t, err)
	assert.Equal(t, challenge.InviteToken.String(), link)
}

func TestGetInviteLink_Forbidden(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	challenge, err := svc.Create(context.Background(), uuid.New(), CreateChallengeParams{
		CategoryID: 1,
		Title:      "Invite Link",
		StartsAt:   futureDate(1),
		EndsAt:     futureDate(30),
	})
	require.NoError(t, err)

	_, err = svc.GetInviteLink(context.Background(), challenge.ID, uuid.New())
	assert.ErrorIs(t, err, ErrForbidden)
}

// --- ListPublic tests ---

func TestListPublic_ClampsLimitAndOffset(t *testing.T) {
	svc, _, _, _ := newTestChallengeService()

	result, err := svc.ListPublic(context.Background(), nil, "", -1, -5)
	require.NoError(t, err)
	assert.Empty(t, result)
}