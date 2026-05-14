package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrNotFound       = errors.New("не найдено")
	ErrForbidden      = errors.New("действие запрещено")
	ErrNotUpcoming    = errors.New("редактировать можно только челлендж, который ещё не начался")
	ErrAlreadyJoined  = errors.New("вы уже присоединились")
	ErrNotPublic      = errors.New("челлендж не публичный")
	ErrChallengeEnded = errors.New("челлендж уже завершён")
)

type CreateChallengeParams struct {
	CategoryID                    int     `json:"category_id"`
	Title                         string  `json:"title"`
	Description                   *string `json:"description"`
	StartsAt                      string  `json:"starts_at"`
	EndsAt                        string  `json:"ends_at"`
	WorkingDays                   []int64 `json:"working_days"`
	MaxSkips                      int     `json:"max_skips"`
	DeadlineTime                  string  `json:"deadline_time"`
	DeadlineTimezoneOffsetMinutes int     `json:"deadline_timezone_offset_minutes"`
	IsPublic                      bool    `json:"is_public"`
}

type UpdateChallengeParams struct {
	CategoryID   *int    `json:"category_id"`
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	StartsAt     *string `json:"starts_at"`
	EndsAt       *string `json:"ends_at"`
	WorkingDays  []int64 `json:"working_days"`
	MaxSkips     *int    `json:"max_skips"`
	DeadlineTime *string `json:"deadline_time"`
	IsPublic     *bool   `json:"is_public"`
}

type ChallengeService struct {
	challenges   ChallengeRepo
	participants ParticipantRepo
	feed         FeedRepo
	badgeSvc     *BadgeService
}

func NewChallengeService(c ChallengeRepo, p ParticipantRepo, f FeedRepo) *ChallengeService {
	return &ChallengeService{challenges: c, participants: p, feed: f}
}

func (s *ChallengeService) SetBadgeService(bs *BadgeService) {
	s.badgeSvc = bs
}

func convertLocalDeadlineToUTC(deadlineTime string, offsetMin int) string {
	parts := strings.Split(deadlineTime, ":")
	if len(parts) < 2 {
		return deadlineTime
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return deadlineTime
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return deadlineTime
	}
	sec := 0
	if len(parts) >= 3 {
		if s, err := strconv.Atoi(parts[2]); err == nil {
			sec = s
		}
	}
	totalLocal := h*60 + m
	totalUTC := totalLocal - offsetMin
	for totalUTC < 0 {
		totalUTC += 1440
	}
	for totalUTC >= 1440 {
		totalUTC -= 1440
	}
	return fmt.Sprintf("%02d:%02d:%02d", totalUTC/60, totalUTC%60, sec)
}

func (s *ChallengeService) Create(ctx context.Context, creatorID uuid.UUID, p CreateChallengeParams) (*domain.Challenge, error) {
	startsAt, err := time.Parse("2006-01-02", p.StartsAt)
	if err != nil {
		return nil, errors.New("неверный формат даты начала, используйте YYYY-MM-DD")
	}
	endsAt, err := time.Parse("2006-01-02", p.EndsAt)
	if err != nil {
		return nil, errors.New("неверный формат даты окончания, используйте YYYY-MM-DD")
	}

	status := "upcoming"
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if !startsAt.After(today) {
		status = "active"
	}

	deadlineTime := p.DeadlineTime
	if deadlineTime == "" {
		deadlineTime = "23:00"
	}
	deadlineTime = convertLocalDeadlineToUTC(deadlineTime, p.DeadlineTimezoneOffsetMinutes)

	workingDays := pq.Int64Array(p.WorkingDays)
	if len(workingDays) == 0 {
		workingDays = pq.Int64Array{0, 1, 2, 3, 4, 5, 6}
	}

	c := &domain.Challenge{
		ID:           uuid.New(),
		CreatorID:    creatorID,
		CategoryID:   p.CategoryID,
		Title:        p.Title,
		Description:  p.Description,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
		WorkingDays:  workingDays,
		MaxSkips:     p.MaxSkips,
		DeadlineTime: deadlineTime,
		IsPublic:     p.IsPublic,
		Status:       status,
	}

	if err := s.challenges.Create(ctx, c); err != nil {
		return nil, err
	}

	_ = s.participants.Add(ctx, c.ID, creatorID)

	_ = s.feed.Insert(ctx, &domain.FeedEvent{
		ID:          uuid.New(),
		ChallengeID: c.ID,
		UserID:      creatorID,
		Type:        "challenge_created",
	})

	return c, nil
}

func (s *ChallengeService) Update(ctx context.Context, challengeID, creatorID uuid.UUID, p UpdateChallengeParams) (*domain.Challenge, error) {
	c, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if c.CreatorID != creatorID {
		return nil, ErrForbidden
	}
	if c.Status != "upcoming" {
		return nil, ErrNotUpcoming
	}

	if p.CategoryID != nil {
		c.CategoryID = *p.CategoryID
	}
	if p.Title != nil {
		c.Title = *p.Title
	}
	if p.Description != nil {
		c.Description = p.Description
	}
	if p.StartsAt != nil {
		t, err := time.Parse("2006-01-02", *p.StartsAt)
		if err != nil {
			return nil, errors.New("неверный формат даты начала")
		}
		c.StartsAt = t
	}
	if p.EndsAt != nil {
		t, err := time.Parse("2006-01-02", *p.EndsAt)
		if err != nil {
			return nil, errors.New("неверный формат даты окончания")
		}
		c.EndsAt = t
	}
	if p.WorkingDays != nil {
		c.WorkingDays = pq.Int64Array(p.WorkingDays)
	}
	if p.MaxSkips != nil {
		c.MaxSkips = *p.MaxSkips
	}
	if p.DeadlineTime != nil {
		c.DeadlineTime = *p.DeadlineTime
	}
	if p.IsPublic != nil {
		c.IsPublic = *p.IsPublic
	}

	if err := s.challenges.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ChallengeService) Finish(ctx context.Context, challengeID, creatorID uuid.UUID) error {
	c, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if c.CreatorID != creatorID {
		return ErrForbidden
	}
	if err := s.challenges.SetStatus(ctx, challengeID, "finished"); err != nil {
		return err
	}
	if s.badgeSvc != nil {
		go s.badgeSvc.ProcessFinishedChallenge(context.Background(), challengeID)
	}
	return nil
}

func (s *ChallengeService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Challenge, error) {
	c, err := s.challenges.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *ChallengeService) IsParticipant(ctx context.Context, challengeID, userID uuid.UUID) bool {
	exists, _ := s.participants.Exists(ctx, challengeID, userID)
	return exists
}

func (s *ChallengeService) ParticipantCount(ctx context.Context, challengeID uuid.UUID) int {
	count, _ := s.participants.Count(ctx, challengeID)
	return count
}

func (s *ChallengeService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.challenges.SetStatus(ctx, id, status)
}

func (s *ChallengeService) ListForUser(ctx context.Context, userID uuid.UUID) ([]domain.Challenge, error) {
	return s.challenges.ListForUser(ctx, userID)
}

func (s *ChallengeService) ListPublic(ctx context.Context, categoryID *int, search string, limit, offset int) ([]domain.Challenge, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.challenges.ListPublic(ctx, categoryID, search, limit, offset)
}

func (s *ChallengeService) JoinByInviteToken(ctx context.Context, userID uuid.UUID, token uuid.UUID) (*domain.Challenge, error) {
	c, err := s.challenges.GetByInviteToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if c.Status == "finished" {
		return nil, ErrChallengeEnded
	}
	joined, _ := s.participants.Exists(ctx, c.ID, userID)
	if joined {
		return c, ErrAlreadyJoined
	}
	if err := s.participants.Add(ctx, c.ID, userID); err != nil {
		return nil, err
	}
	_ = s.feed.Insert(ctx, &domain.FeedEvent{
		ID:          uuid.New(),
		ChallengeID: c.ID,
		UserID:      userID,
		Type:        "user_joined",
	})
	return c, nil
}

func (s *ChallengeService) JoinPublic(ctx context.Context, userID, challengeID uuid.UUID) error {
	c, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if !c.IsPublic {
		return ErrNotPublic
	}
	if c.Status == "finished" {
		return ErrChallengeEnded
	}
	joined, _ := s.participants.Exists(ctx, c.ID, userID)
	if joined {
		return ErrAlreadyJoined
	}
	if err := s.participants.Add(ctx, c.ID, userID); err != nil {
		return err
	}
	_ = s.feed.Insert(ctx, &domain.FeedEvent{
		ID:          uuid.New(),
		ChallengeID: c.ID,
		UserID:      userID,
		Type:        "user_joined",
	})
	return nil
}

func (s *ChallengeService) GetInviteLink(ctx context.Context, challengeID, requesterID uuid.UUID) (string, error) {
	c, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if c.CreatorID != requesterID {
		isParticipant, _ := s.participants.Exists(ctx, challengeID, requesterID)
		if !isParticipant {
			return "", ErrForbidden
		}
	}
	return c.InviteToken.String(), nil
}

func (s *ChallengeService) RemoveParticipant(ctx context.Context, challengeID, creatorID, targetUserID uuid.UUID) error {
	c, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if c.CreatorID != creatorID {
		return ErrForbidden
	}
	if targetUserID == creatorID {
		return errors.New("автор не может удалить сам себя из челленджа")
	}
	return s.participants.Remove(ctx, challengeID, targetUserID)
}
