package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math"
	"time"

	"tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrNotParticipant     = errors.New("вы не являетесь участником этого челленджа")
	ErrNotWorkingDay      = errors.New("сегодня не рабочий день в этом челлендже")
	ErrAlreadyChecked     = errors.New("вы уже отметились сегодня")
	ErrNotCheckedIn       = errors.New("вы ещё не отметились сегодня")
	ErrChallengeNotActive = errors.New("челлендж не активен")
	ErrDeadlinePassed     = errors.New("время для отметки на сегодня уже истекло")
	ErrUndoNotAllowed     = errors.New("отменить отметку можно только до дедлайна")
)

func deadlineTodayUTC(deadlineTime string, now time.Time) (time.Time, error) {
	if deadlineTime == "" {
		return time.Time{}, errors.New("empty deadline_time")
	}
	layouts := []string{
		"15:04:05",
		"15:04",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05 -0700 MST",
	}
	var t time.Time
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, deadlineTime)
		if err == nil {
			t = parsed
			lastErr = nil
			break
		}
		lastErr = err
	}
	if lastErr != nil {
		return time.Time{}, fmt.Errorf("parse deadline_time %q: %w", deadlineTime, lastErr)
	}
	n := now.UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
}

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

func (s *CheckInService) SetBadgeService(bs *BadgeService) {
	s.badgeSvc = bs
}

func (s *CheckInService) CheckIn(ctx context.Context, userID, challengeID uuid.UUID, comment, imageURL string) (*domain.SimpleCheckIn, error) {
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

	now := time.Now().UTC()
	log.Printf("DEADLINE CHECK: now=%s deadline_raw=%s",
		time.Now().UTC(), challenge.DeadlineTime)
	rawDB, rawErr := s.challenges.GetDeadlineTimeText(ctx, challengeID)
	fmt.Printf("DEBUG deadline_time field: %q, deadline_time from DB: %q (err=%v), now: %s\n",
		challenge.DeadlineTime, rawDB, rawErr, now.Format(time.RFC3339))

	rawDeadline := challenge.DeadlineTime
	if rawErr == nil && rawDB != "" {
		rawDeadline = rawDB
	}
	deadline, derr := deadlineTodayUTC(rawDeadline, now)
	slog.Info("checkin deadline check",
		"challenge_id", challengeID,
		"raw_deadline_time", challenge.DeadlineTime,
		"raw_from_db", rawDB,
		"parsed_deadline", deadline,
		"now_utc", now,
		"parse_err", derr,
	)
	if derr != nil {
		// Если поле не парсится — лучше отказать, чем тихо пропустить проверку.
		return nil, fmt.Errorf("invalid deadline_time: %w", derr)
	}
	if !now.Before(deadline) {
		return nil, ErrDeadlinePassed
	}

	exists, err := s.checkIns.ExistsForDate(ctx, challengeID, userID, today)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyChecked
	}

	var imageURLPtr *string
	if imageURL != "" {
		imageURLPtr = &imageURL
	}
	ci := &domain.SimpleCheckIn{
		ID:          uuid.New(),
		ChallengeID: challengeID,
		UserID:      userID,
		Date:        today,
		Comment:     comment,
		ImageURL:    imageURLPtr,
	}

	if err := s.checkIns.Create(ctx, ci); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyChecked
		}
		return nil, fmt.Errorf("create check-in: %w", err)
	}

	allCheckIns, _ := s.checkIns.ListForUser(ctx, challengeID, userID)
	dayNumber := len(allCheckIns)

	doneMap := make(map[string]bool)
	for _, c := range allCheckIns {
		doneMap[normalizeDate(c.Date).Format("2006-01-02")] = true
	}
	wdSet := make(map[int]bool)
	for _, wd := range challenge.WorkingDays {
		wdSet[int(wd)] = true
	}
	currentStreak, _ := computeStreaksSimple(doneMap, challenge.StartsAt, challenge.EndsAt, wdSet)

	feedPayload := map[string]any{
		"comment":    comment,
		"day_number": dayNumber,
		"streak":     currentStreak,
	}
	if imageURL != "" {
		feedPayload["image_url"] = imageURL
	}
	feedData, _ := json.Marshal(feedPayload)
	rawData := json.RawMessage(feedData)

	refID := ci.ID
	_ = s.feed.Insert(ctx, &domain.FeedEvent{
		ID:          uuid.New(),
		ChallengeID: challengeID,
		UserID:      userID,
		Type:        "check_in",
		ReferenceID: &refID,
		Data:        &rawData,
	})

	if s.badgeSvc != nil {
		go s.badgeSvc.CheckAndAward(context.Background(), userID, challengeID)
	}

	return ci, nil
}

func (s *CheckInService) Undo(ctx context.Context, userID, challengeID uuid.UUID) error {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)

	ci, err := s.checkIns.GetForDate(ctx, challengeID, userID, today)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotCheckedIn
		}
		return err
	}

	deadline, derr := deadlineTodayUTC(challenge.DeadlineTime, now)
	if derr != nil {
		return fmt.Errorf("invalid deadline_time: %w", derr)
	}
	if !now.Before(deadline) {
		return ErrUndoNotAllowed
	}

	if err := s.checkIns.Delete(ctx, challengeID, userID, today); err != nil {
		return err
	}

	_ = s.feed.DeleteByReference(ctx, ci.ID, "check_in")
	return nil
}

func (s *CheckInService) GetProgress(ctx context.Context, userID, challengeID uuid.UUID) (*domain.Progress, error) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	dayIdx := (int(today.Weekday()) + 6) % 7
	isWorking := false
	for _, wd := range challenge.WorkingDays {
		if int(wd) == dayIdx {
			isWorking = true
			break
		}
	}

	checkedInToday, err := s.checkIns.ExistsForDate(ctx, challengeID, userID, today)
	if err != nil {
		return nil, err
	}

	checkIns, err := s.checkIns.ListForUser(ctx, challengeID, userID)
	if err != nil {
		return nil, err
	}

	doneDays := len(checkIns)

	doneMap := make(map[string]bool)
	for _, ci := range checkIns {
		doneMap[normalizeDate(ci.Date).Format("2006-01-02")] = true
	}

	wdSet := make(map[int]bool)
	for _, wd := range challenge.WorkingDays {
		wdSet[int(wd)] = true
	}

	totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)

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

func (s *CheckInService) ListAll(ctx context.Context, challengeID, userID uuid.UUID) ([]domain.SimpleCheckIn, error) {
	return s.checkIns.ListForUser(ctx, challengeID, userID)
}
