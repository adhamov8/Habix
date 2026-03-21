package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"tracker/internal/domain"
)

type BadgeService struct {
	badges       BadgeRepo
	checkIns     CheckInRepo
	challenges   ChallengeRepo
	participants ParticipantRepo
	feed         FeedRepo
	notifSvc     *NotificationService
}

func NewBadgeService(
	b BadgeRepo,
	ci CheckInRepo,
	ch ChallengeRepo,
	p ParticipantRepo,
	f FeedRepo,
) *BadgeService {
	return &BadgeService{badges: b, checkIns: ci, challenges: ch, participants: p, feed: f}
}

func (s *BadgeService) SetNotificationService(ns *NotificationService) {
	s.notifSvc = ns
}

// CheckAndAward checks all badge conditions after a check-in and awards earned badges.
func (s *BadgeService) CheckAndAward(ctx context.Context, userID, challengeID uuid.UUID) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		return
	}

	// Get user check-ins for this challenge
	checkIns, err := s.checkIns.ListForUser(ctx, challengeID, userID)
	if err != nil {
		return
	}

	doneDays := len(checkIns)

	// Build done date set and compute streaks
	doneMap := make(map[string]bool)
	for _, ci := range checkIns {
		doneMap[normalizeDate(ci.Date).Format("2006-01-02")] = true
	}

	wdSet := make(map[int]bool)
	for _, wd := range challenge.WorkingDays {
		wdSet[int(wd)] = true
	}

	currentStreak, _ := computeStreaksSimple(doneMap, challenge.StartsAt, challenge.EndsAt, wdSet)

	// 1. First check-in
	if doneDays == 1 {
		s.awardBadge(ctx, userID, "first_checkin", &challengeID)
	}

	// 2. Streak badges
	if currentStreak >= 3 {
		s.awardBadge(ctx, userID, "streak_3", &challengeID)
	}
	if currentStreak >= 7 {
		s.awardBadge(ctx, userID, "streak_7", &challengeID)
	}
	if currentStreak >= 30 {
		s.awardBadge(ctx, userID, "streak_30", &challengeID)
	}

	// 3. Perfect week — check if all working days in the current calendar week are done
	if s.isPerfectWeek(doneMap, wdSet) {
		s.awardBadge(ctx, userID, "perfect_week", &challengeID)
	}

	// 4. Challenge complete — 100% adherence on a finished challenge
	if challenge.Status == "finished" {
		totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)
		if totalWorkingDays > 0 && doneDays >= totalWorkingDays {
			s.awardBadge(ctx, userID, "challenge_complete", &challengeID)
		}
	}

	// 5. Join 3 challenges
	count, err := s.badges.CountUserChallenges(ctx, userID)
	if err == nil && count >= 3 {
		s.awardBadge(ctx, userID, "join_3_challenges", nil)
	}
}

func (s *BadgeService) awardBadge(ctx context.Context, userID uuid.UUID, code string, challengeID *uuid.UUID) {
	awarded, err := s.badges.Award(ctx, userID, code, challengeID)
	if err != nil || !awarded {
		return
	}

	// Get badge definition for title/icon
	bd, err := s.badges.GetDefinitionByCode(ctx, code)
	if err != nil || bd == nil {
		return
	}

	// Insert feed event if challenge-specific
	if challengeID != nil {
		feedData, _ := json.Marshal(map[string]any{
			"badge_title": bd.Title,
			"badge_icon":  bd.Icon,
			"badge_code":  code,
		})
		rawData := json.RawMessage(feedData)
		_ = s.feed.Insert(ctx, &domain.FeedEvent{
			ID:          uuid.New(),
			ChallengeID: *challengeID,
			UserID:      userID,
			Type:        "badge_earned",
			Data:        &rawData,
		})
	}

	// Send notification
	if s.notifSvc != nil {
		_ = s.notifSvc.Notify(ctx, userID, "badge_earned",
			bd.Icon+" Новое достижение!",
			"Вы получили достижение «"+bd.Title+"»",
			map[string]interface{}{"badge_code": code},
		)
	}
}

// isPerfectWeek checks if all working days in the current ISO week are checked in.
func (s *BadgeService) isPerfectWeek(doneMap map[string]bool, wdSet map[int]bool) bool {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	// Find Monday of this week
	daysSinceMonday := (int(today.Weekday()) + 6) % 7
	monday := today.AddDate(0, 0, -daysSinceMonday)

	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		if d.After(today) {
			break // Can't check future days
		}
		dayIdx := (int(d.Weekday()) + 6) % 7
		if wdSet[dayIdx] {
			if !doneMap[d.Format("2006-01-02")] {
				return false
			}
		}
	}
	return true
}

func (s *BadgeService) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	return s.badges.ListForUser(ctx, userID)
}

func (s *BadgeService) ListDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error) {
	return s.badges.ListDefinitions(ctx)
}