package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"tracker/internal/domain"
	"tracker/internal/repository"
)

// --- UserRepo mock ---

type mockUserRepo struct {
	users map[string]*domain.User // keyed by email
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return u, nil
}

func (m *mockUserRepo) Create(_ context.Context, u *domain.User) error {
	m.users[u.Email] = u
	return nil
}

// --- TokenRepo mock ---

type mockTokenRepo struct {
	tokens map[string]*repository.RefreshToken // keyed by hash
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{tokens: make(map[string]*repository.RefreshToken)}
}

func (m *mockTokenRepo) Create(_ context.Context, t *repository.RefreshToken) error {
	m.tokens[t.TokenHash] = t
	return nil
}

func (m *mockTokenRepo) GetByHash(_ context.Context, hash string) (*repository.RefreshToken, error) {
	t, ok := m.tokens[hash]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return t, nil
}

func (m *mockTokenRepo) Delete(_ context.Context, hash string) error {
	delete(m.tokens, hash)
	return nil
}

// --- ChallengeRepo mock ---

type mockChallengeRepo struct {
	challenges map[uuid.UUID]*domain.Challenge
}

func newMockChallengeRepo() *mockChallengeRepo {
	return &mockChallengeRepo{challenges: make(map[uuid.UUID]*domain.Challenge)}
}

func (m *mockChallengeRepo) Create(_ context.Context, c *domain.Challenge) error {
	m.challenges[c.ID] = c
	return nil
}

func (m *mockChallengeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Challenge, error) {
	c, ok := m.challenges[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return c, nil
}

func (m *mockChallengeRepo) GetByInviteToken(_ context.Context, token uuid.UUID) (*domain.Challenge, error) {
	for _, c := range m.challenges {
		if c.InviteToken == token {
			return c, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockChallengeRepo) Update(_ context.Context, c *domain.Challenge) error {
	m.challenges[c.ID] = c
	return nil
}

func (m *mockChallengeRepo) SetStatus(_ context.Context, id uuid.UUID, status string) error {
	c, ok := m.challenges[id]
	if !ok {
		return sql.ErrNoRows
	}
	c.Status = status
	return nil
}

func (m *mockChallengeRepo) ListForUser(_ context.Context, _ uuid.UUID) ([]domain.Challenge, error) {
	return nil, nil
}

func (m *mockChallengeRepo) ListPublic(_ context.Context, _ *int, _ string, _, _ int) ([]domain.Challenge, error) {
	return nil, nil
}

// --- ParticipantRepo mock ---

type mockParticipantRepo struct {
	participants map[string]bool // "challengeID:userID"
}

func newMockParticipantRepo() *mockParticipantRepo {
	return &mockParticipantRepo{participants: make(map[string]bool)}
}

func (m *mockParticipantRepo) key(cid, uid uuid.UUID) string {
	return fmt.Sprintf("%s:%s", cid, uid)
}

func (m *mockParticipantRepo) Add(_ context.Context, cid, uid uuid.UUID) error {
	m.participants[m.key(cid, uid)] = true
	return nil
}

func (m *mockParticipantRepo) Remove(_ context.Context, cid, uid uuid.UUID) error {
	delete(m.participants, m.key(cid, uid))
	return nil
}

func (m *mockParticipantRepo) Exists(_ context.Context, cid, uid uuid.UUID) (bool, error) {
	return m.participants[m.key(cid, uid)], nil
}

func (m *mockParticipantRepo) Count(_ context.Context, cid uuid.UUID) (int, error) {
	count := 0
	prefix := cid.String() + ":"
	for k := range m.participants {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			count++
		}
	}
	return count, nil
}

func (m *mockParticipantRepo) ListByChallenge(_ context.Context, _ uuid.UUID) ([]domain.Participant, error) {
	return nil, nil
}

// --- FeedRepo mock ---

type mockFeedRepo struct {
	events []*domain.FeedEvent
}

func newMockFeedRepo() *mockFeedRepo {
	return &mockFeedRepo{}
}

func (m *mockFeedRepo) Insert(_ context.Context, e *domain.FeedEvent) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockFeedRepo) ListByChallenge(_ context.Context, _ uuid.UUID, _, _ int) ([]domain.FeedEvent, error) {
	return nil, nil
}

// --- CheckInRepo mock ---

type mockCheckInRepo struct {
	checkIns     []domain.SimpleCheckIn
	existsResult bool
	existsErr    error
	createErr    error
	created      []*domain.SimpleCheckIn
}

func (m *mockCheckInRepo) Create(_ context.Context, ci *domain.SimpleCheckIn) error {
	m.created = append(m.created, ci)
	return m.createErr
}

func (m *mockCheckInRepo) Delete(_ context.Context, _, _ uuid.UUID, _ time.Time) error {
	return nil
}

func (m *mockCheckInRepo) ExistsForDate(_ context.Context, _, _ uuid.UUID, _ time.Time) (bool, error) {
	return m.existsResult, m.existsErr
}

func (m *mockCheckInRepo) ListForUser(_ context.Context, _, _ uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return m.checkIns, nil
}

func (m *mockCheckInRepo) ListForChallenge(_ context.Context, _ uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return nil, nil
}

// --- BadgeRepo mock ---

type mockBadgeRepo struct {
	awarded     map[string]bool
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
		return false, nil
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