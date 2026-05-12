package service

import (
	"context"
	"encoding/json"
	"time"

	"tracker/internal/domain"

	"github.com/google/uuid"
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

// проверяем условия значков после отметки и выдаёт те, которые заслужил пользователь
func (s *BadgeService) CheckAndAward(ctx context.Context, userID, challengeID uuid.UUID) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		return
	}

	checkIns, err := s.checkIns.ListForUser(ctx, challengeID, userID)
	if err != nil {
		return
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

	currentStreak, _ := computeStreaksSimple(doneMap, challenge.StartsAt, challenge.EndsAt, wdSet)

	if doneDays == 1 {
		s.awardBadge(ctx, userID, "first_checkin", &challengeID)
	}

	if currentStreak >= 3 {
		s.awardBadge(ctx, userID, "streak_3", &challengeID)
	}
	if currentStreak >= 7 {
		s.awardBadge(ctx, userID, "streak_7", &challengeID)
	}
	if currentStreak >= 30 {
		s.awardBadge(ctx, userID, "streak_30", &challengeID)
	}

	if s.isPerfectWeek(doneMap, wdSet) {
		s.awardBadge(ctx, userID, "perfect_week", &challengeID)
	}

	// На завершённом челлендже выдаём значок только при 100% выполнении
	if challenge.Status == "finished" {
		totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)
		if totalWorkingDays > 0 && doneDays >= totalWorkingDays {
			s.awardBadge(ctx, userID, "challenge_complete", &challengeID)
		}
	}

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

	bd, err := s.badges.GetDefinitionByCode(ctx, code)
	if err != nil || bd == nil {
		return
	}

	// Событие в ленту пишем только если значок привязан к конкретному челленджу
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

	if s.notifSvc != nil {
		_ = s.notifSvc.Notify(ctx, userID, "badge_earned",
			bd.Icon+" Новое достижение!",
			"Вы получили достижение «"+bd.Title+"»",
			map[string]interface{}{"badge_code": code},
		)
	}
}

// проверяем, отмечены ли все рабочие дни текущей недели
func (s *BadgeService) isPerfectWeek(doneMap map[string]bool, wdSet map[int]bool) bool {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	// Находим понедельник текущей недели
	daysSinceMonday := (int(today.Weekday()) + 6) % 7
	monday := today.AddDate(0, 0, -daysSinceMonday)

	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		if d.After(today) {
			break // Будущие дни пропускаем
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

// выдаём значки за завершение и значок «Ветеран»
// одному участнику после окончания челленджа
func (s *BadgeService) CheckAndAwardOnFinish(ctx context.Context, userID, challengeID uuid.UUID, adherencePct float64) {
	if adherencePct >= 50 {
		s.awardBadge(ctx, userID, "complete_50", &challengeID)
	}
	if adherencePct >= 80 {
		s.awardBadge(ctx, userID, "complete_80", &challengeID)
	}
	if adherencePct >= 100 {
		s.awardBadge(ctx, userID, "challenge_complete", &challengeID)
	}
	if finished, err := s.badges.CountFinishedChallengesForUser(ctx, userID); err == nil && finished >= 5 {
		s.awardBadge(ctx, userID, "veteran", nil)
	}
}

// считаем выполнение каждого участника
// только что завершённого челленджа и выдаёт нужные значки
func (s *BadgeService) ProcessFinishedChallenge(ctx context.Context, challengeID uuid.UUID) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		return
	}
	totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)
	if totalWorkingDays <= 0 {
		return
	}
	participants, err := s.participants.ListByChallenge(ctx, challengeID)
	if err != nil {
		return
	}
	for _, p := range participants {
		checkIns, err := s.checkIns.ListForUser(ctx, challengeID, p.UserID)
		if err != nil {
			continue
		}
		adherence := float64(len(checkIns)) / float64(totalWorkingDays) * 100
		s.CheckAndAwardOnFinish(ctx, p.UserID, challengeID, adherence)
	}
}

func (s *BadgeService) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	return s.badges.ListForUser(ctx, userID)
}

func (s *BadgeService) ListDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error) {
	return s.badges.ListDefinitions(ctx)
}
