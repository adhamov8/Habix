package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"tracker/internal/domain"
)

var (
	ErrNotParticipant     = errors.New("user is not a participant")
	ErrNotWorkingDay      = errors.New("today is not a working day for this challenge")
	ErrAlreadyChecked     = errors.New("already checked in today")
	ErrNotCheckedIn       = errors.New("not checked in today")
	ErrChallengeNotActive = errors.New("challenge is not active")
)

type CheckInService struct {
	checkIns     CheckInRepo
	challenges   ChallengeRepo
	participants ParticipantRepo
	feed         FeedRepo
	badgeSvc     *BadgeService
}

func NewCheckInService(
	ci CheckInRepo,
	ch ChallengeRepo,
	p ParticipantRepo,
	f FeedRepo,
) *CheckInService {
	return &CheckInService{checkIns: ci, challenges: ch, participants: p, feed: f}
}

// SetBadgeService sets the badge service for awarding badges after check-ins.
func (s *CheckInService) SetBadgeService(bs *BadgeService) {
	s.badgeSvc = bs
}

// CheckIn creates a check-in for today with an optional comment.
func (s *CheckInService) CheckIn(ctx context.Context, userID, challengeID uuid.UUID, comment string) (*domain.SimpleCheckIn, error) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if challenge.Status != "active" {
		return nil, ErrChallengeNotActive
	}

	isParticipant, err := s.participants.Exists(ctx, challengeID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Check working day
	dayIdx := (int(today.Weekday()) + 6) % 7
	isWorking := false
	for _, wd := range challenge.WorkingDays {
		if int(wd) == dayIdx {
			isWorking = true
			break
		}
	}
	if !isWorking {
		return nil, ErrNotWorkingDay
	}

	// Check not already checked in
	exists, err := s.checkIns.ExistsForDate(ctx, challengeID, userID, today)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyChecked
	}

	ci := &domain.SimpleCheckIn{
		ID:          uuid.New(),
		ChallengeID: challengeID,
		UserID:      userID,
		Date:        today,
		Comment:     comment,
	}

	if err := s.checkIns.Create(ctx, ci); err != nil {
		return nil, fmt.Errorf("create check-in: %w", err)
	}

	// Build feed event data with comment, day number and streak
	allCheckIns, _ := s.checkIns.ListForUser(ctx, challengeID, userID)
	dayNumber := len(allCheckIns)

	// Compute current streak for feed display
	doneMap := make(map[string]bool)
	for _, c := range allCheckIns {
		doneMap[normalizeDate(c.Date).Format("2006-01-02")] = true
	}
	wdSet := make(map[int]bool)
	for _, wd := range challenge.WorkingDays {
		wdSet[int(wd)] = true
	}
	currentStreak, _ := computeStreaksSimple(doneMap, challenge.StartsAt, challenge.EndsAt, wdSet)

	feedData, _ := json.Marshal(map[string]any{
		"comment":    comment,
		"day_number": dayNumber,
		"streak":     currentStreak,
	})
	rawData := json.RawMessage(feedData)

	// Insert feed event
	refID := ci.ID
	_ = s.feed.Insert(ctx, &domain.FeedEvent{
		ID:          uuid.New(),
		ChallengeID: challengeID,
		UserID:      userID,
		Type:        "check_in",
		ReferenceID: &refID,
		Data:        &rawData,
	})

	// Check and award badges
	if s.badgeSvc != nil {
		go s.badgeSvc.CheckAndAward(context.Background(), userID, challengeID)
	}

	return ci, nil
}

// Undo removes today's check-in.
func (s *CheckInService) Undo(ctx context.Context, userID, challengeID uuid.UUID) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	exists, err := s.checkIns.ExistsForDate(ctx, challengeID, userID, today)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotCheckedIn
	}

	return s.checkIns.Delete(ctx, challengeID, userID, today)
}

// GetProgress returns the user's progress in a challenge.
func (s *CheckInService) GetProgress(ctx context.Context, userID, challengeID uuid.UUID) (*domain.Progress, error) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Is today a working day?
	dayIdx := (int(today.Weekday()) + 6) % 7
	isWorking := false
	for _, wd := range challenge.WorkingDays {
		if int(wd) == dayIdx {
			isWorking = true
			break
		}
	}

	// Checked in today?
	checkedInToday, err := s.checkIns.ExistsForDate(ctx, challengeID, userID, today)
	if err != nil {
		return nil, err
	}

	// All check-ins for streak computation
	checkIns, err := s.checkIns.ListForUser(ctx, challengeID, userID)
	if err != nil {
		return nil, err
	}

	doneDays := len(checkIns)

	// Build done date set
	doneMap := make(map[string]bool)
	for _, ci := range checkIns {
		doneMap[normalizeDate(ci.Date).Format("2006-01-02")] = true
	}

	// Working days set
	wdSet := make(map[int]bool)
	for _, wd := range challenge.WorkingDays {
		wdSet[int(wd)] = true
	}

	totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)

	// Compute streaks
	cur, mx := computeStreaksSimple(doneMap, challenge.StartsAt, challenge.EndsAt, wdSet)

	var adherence float64
	if totalWorkingDays > 0 {
		adherence = math.Round(float64(doneDays)/float64(totalWorkingDays)*10000) / 100
	}

	return &domain.Progress{
		CheckedInToday: checkedInToday,
		IsWorkingDay:   isWorking,
		CurrentStreak:  cur,
		MaxStreak:      mx,
		DoneDays:       doneDays,
		TotalDays:      totalWorkingDays,
		AdherencePct:   adherence,
	}, nil
}

// ListAll returns all check-ins for a user in a challenge.
func (s *CheckInService) ListAll(ctx context.Context, challengeID, userID uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return s.checkIns.ListForUser(ctx, challengeID, userID)
}