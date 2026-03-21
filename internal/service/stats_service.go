package service

import (
	"context"
	"errors"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"tracker/internal/domain"
	"tracker/internal/repository"
)

type StatsService struct {
	stats      *repository.StatsRepository
	challenges *repository.ChallengeRepository
}

func NewStatsService(s *repository.StatsRepository, c *repository.ChallengeRepository) *StatsService {
	return &StatsService{stats: s, challenges: c}
}

func (s *StatsService) Leaderboard(ctx context.Context, challengeID uuid.UUID) ([]domain.LeaderboardEntry, error) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		return nil, ErrNotFound
	}

	totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)

	counts, err := s.stats.GetParticipantCounts(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	checkIns, err := s.stats.GetCheckInsForChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	// Group check-in dates by user
	userDoneMaps := make(map[uuid.UUID]map[string]bool)
	for _, ci := range checkIns {
		if userDoneMaps[ci.UserID] == nil {
			userDoneMaps[ci.UserID] = make(map[string]bool)
		}
		userDoneMaps[ci.UserID][normalizeDate(ci.Date).Format("2006-01-02")] = true
	}

	workingDaySet := make(map[int]bool)
	for _, wd := range challenge.WorkingDays {
		workingDaySet[int(wd)] = true
	}

	entries := make([]domain.LeaderboardEntry, 0, len(counts))
	for _, pc := range counts {
		var adherence float64
		if totalWorkingDays > 0 {
			adherence = math.Round(float64(pc.DoneDays)/float64(totalWorkingDays)*10000) / 100
		}

		doneMap := userDoneMaps[pc.UserID]
		if doneMap == nil {
			doneMap = make(map[string]bool)
		}
		cur, mx := computeStreaksSimple(doneMap, challenge.StartsAt, challenge.EndsAt, workingDaySet)

		entries = append(entries, domain.LeaderboardEntry{
			UserID:           pc.UserID,
			UserName:         pc.UserName,
			TotalWorkingDays: totalWorkingDays,
			DoneDays:         pc.DoneDays,
			MissedDays:       0,
			AdherencePct:     adherence,
			CurrentStreak:    cur,
			MaxStreak:        mx,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].AdherencePct != entries[j].AdherencePct {
			return entries[i].AdherencePct > entries[j].AdherencePct
		}
		return entries[i].CurrentStreak > entries[j].CurrentStreak
	})

	return entries, nil
}

func (s *StatsService) PersonalStats(ctx context.Context, userID uuid.UUID) (*domain.PersonalStats, error) {
	challengeStats, err := s.stats.GetUserChallengeStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	ps := &domain.PersonalStats{}
	var totalAdherence float64
	var adherenceCount int

	for _, cs := range challengeStats {
		ps.TotalChallenges++
		switch cs.Status {
		case "active":
			ps.ActiveChallenges++
		case "finished":
			ps.FinishedChallenges++
		}
		if cs.TotalDays > 0 {
			totalAdherence += float64(cs.DoneDays) / float64(cs.TotalDays) * 100
			adherenceCount++
		}
	}

	if adherenceCount > 0 {
		ps.AvgAdherencePct = math.Round(totalAdherence/float64(adherenceCount)*100) / 100
	}

	// Compute max streak across all challenges
	allCheckIns, err := s.stats.GetUserAllCheckIns(ctx, userID)
	if err != nil {
		return ps, nil
	}

	// Simple max streak: consecutive days with check-ins
	maxStreak := 0
	streak := 0
	var prevDate time.Time
	for _, ci := range allCheckIns {
		d := normalizeDate(ci.Date)
		if !prevDate.IsZero() && d.Sub(prevDate) == 24*time.Hour {
			streak++
		} else {
			streak = 1
		}
		if streak > maxStreak {
			maxStreak = streak
		}
		prevDate = d
	}
	ps.MaxStreak = maxStreak

	return ps, nil
}

func (s *StatsService) ChallengeStats(ctx context.Context, challengeID uuid.UUID) (*domain.ChallengeStats, error) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		return nil, ErrNotFound
	}

	participation, err := s.stats.GetParticipationByDay(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)
	counts, err := s.stats.GetParticipantCounts(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	buckets := []domain.AdherenceBucket{
		{Bucket: "0-25%", Count: 0},
		{Bucket: "25-50%", Count: 0},
		{Bucket: "50-75%", Count: 0},
		{Bucket: "75-100%", Count: 0},
	}

	for _, pc := range counts {
		var pct float64
		if totalWorkingDays > 0 {
			pct = float64(pc.DoneDays) / float64(totalWorkingDays) * 100
		}
		switch {
		case pct < 25:
			buckets[0].Count++
		case pct < 50:
			buckets[1].Count++
		case pct < 75:
			buckets[2].Count++
		default:
			buckets[3].Count++
		}
	}

	return &domain.ChallengeStats{
		ParticipationByDay: participation,
		Distribution:       buckets,
	}, nil
}

func (s *StatsService) ChallengeSummary(ctx context.Context, challengeID uuid.UUID) (*domain.ChallengeSummary, error) {
	challenge, err := s.challenges.GetByID(ctx, challengeID)
	if err != nil {
		return nil, ErrNotFound
	}
	if challenge.Status != "finished" {
		return nil, errors.New("challenge is not finished")
	}

	totalWorkingDays := countWorkingDays(challenge.StartsAt, challenge.EndsAt, challenge.WorkingDays)

	counts, err := s.stats.GetParticipantCounts(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	checkIns, err := s.stats.GetCheckInsForChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	// Group check-in dates by user
	userDoneMaps := make(map[uuid.UUID]map[string]bool)
	for _, ci := range checkIns {
		if userDoneMaps[ci.UserID] == nil {
			userDoneMaps[ci.UserID] = make(map[string]bool)
		}
		userDoneMaps[ci.UserID][normalizeDate(ci.Date).Format("2006-01-02")] = true
	}

	workingDaySet := make(map[int]bool)
	for _, wd := range challenge.WorkingDays {
		workingDaySet[int(wd)] = true
	}

	summary := &domain.ChallengeSummary{
		TotalParticipants: len(counts),
		TotalCheckIns:     len(checkIns),
		Participants:      make([]domain.ParticipantDetail, 0, len(counts)),
	}

	var totalAdherence float64
	var bestAdherence float64

	for _, pc := range counts {
		var adherence float64
		if totalWorkingDays > 0 {
			adherence = math.Round(float64(pc.DoneDays)/float64(totalWorkingDays)*10000) / 100
		}

		doneMap := userDoneMaps[pc.UserID]
		if doneMap == nil {
			doneMap = make(map[string]bool)
		}
		_, mx := computeStreaksSimple(doneMap, challenge.StartsAt, challenge.EndsAt, workingDaySet)

		pd := domain.ParticipantDetail{
			UserID:    pc.UserID,
			UserName:  pc.UserName,
			Adherence: adherence,
			MaxStreak: mx,
			DoneDays:  pc.DoneDays,
		}

		summary.Participants = append(summary.Participants, pd)
		totalAdherence += adherence

		if adherence >= 80 {
			summary.ParticipantsFinished++
		}
		if adherence > bestAdherence {
			bestAdherence = adherence
			best := pd
			summary.BestPerformer = &best
		}
	}

	if len(counts) > 0 {
		summary.AvgAdherence = math.Round(totalAdherence/float64(len(counts))*100) / 100
	}

	// Sort participants by adherence descending
	sort.Slice(summary.Participants, func(i, j int) bool {
		return summary.Participants[i].Adherence > summary.Participants[j].Adherence
	})

	return summary, nil
}

// --- helpers ---

// normalizeDate extracts year/month/day and returns midnight UTC,
// regardless of the timezone lib/pq attached to a DATE column.
func normalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func countWorkingDays(startsAt, endsAt time.Time, workingDays []int64) int {
	wdSet := make(map[int]bool)
	for _, wd := range workingDays {
		wdSet[int(wd)] = true
	}

	start := normalizeDate(startsAt)
	end := normalizeDate(endsAt)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if end.After(today) {
		end = today
	}

	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayIdx := (int(d.Weekday()) + 6) % 7
		if wdSet[dayIdx] {
			count++
		}
	}
	return count
}

// computeStreaksSimple computes current and max streaks from a set of done dates.
func computeStreaksSimple(doneMap map[string]bool, startsAt, endsAt time.Time, workingDaySet map[int]bool) (current, max int) {
	start := normalizeDate(startsAt)
	end := normalizeDate(endsAt)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if end.After(today) {
		end = today
	}

	// Walk working days from start to end
	var workingDayList []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayIdx := (int(d.Weekday()) + 6) % 7
		if workingDaySet[dayIdx] {
			workingDayList = append(workingDayList, d)
		}
	}

	streak := 0
	for _, d := range workingDayList {
		if doneMap[d.Format("2006-01-02")] {
			streak++
			if streak > max {
				max = streak
			}
		} else {
			streak = 0
		}
	}
	current = streak

	return current, max
}