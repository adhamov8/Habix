package domain

import "github.com/google/uuid"

type LeaderboardEntry struct {
	UserID           uuid.UUID `json:"user_id"`
	UserName         string    `json:"user_name"`
	TotalWorkingDays int       `json:"total_working_days"`
	DoneDays         int       `json:"done_days"`
	MissedDays       int       `json:"missed_days"`
	AdherencePct     float64   `json:"adherence_pct"`
	CurrentStreak    int       `json:"current_streak"`
	MaxStreak        int       `json:"max_streak"`
}

type PersonalStats struct {
	TotalChallenges    int     `json:"total_challenges"`
	ActiveChallenges   int     `json:"active_challenges"`
	FinishedChallenges int     `json:"finished_challenges"`
	AvgAdherencePct    float64 `json:"avg_adherence_pct"`
	MaxStreak          int     `json:"max_streak"`
}

type DayParticipation struct {
	Date       string `db:"date"        json:"date"`
	CheckedIn  int    `db:"checked_in"  json:"checked_in"`
	TotalUsers int    `db:"total_users" json:"total_users"`
}

type AdherenceBucket struct {
	Bucket string `json:"bucket"`
	Count  int    `json:"count"`
}

type ChallengeStats struct {
	ParticipationByDay []DayParticipation `json:"participation_by_day"`
	Distribution       []AdherenceBucket  `json:"distribution"`
}

type ParticipantDetail struct {
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	Adherence float64   `json:"adherence"`
	MaxStreak int       `json:"max_streak"`
	DoneDays  int       `json:"done_days"`
}

type ChallengeSummary struct {
	TotalParticipants    int                 `json:"total_participants"`
	AvgAdherence         float64             `json:"avg_adherence"`
	BestPerformer        *ParticipantDetail  `json:"best_performer"`
	TotalCheckIns        int                 `json:"total_checkins"`
	ParticipantsFinished int                 `json:"participants_finished"`
	Participants         []ParticipantDetail `json:"participants"`
}