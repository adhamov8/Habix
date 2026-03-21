package domain

import (
	"time"

	"github.com/google/uuid"
)

// Legacy types kept for old check_ins table (comments/likes still reference it)
type CheckIn struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	ChallengeID uuid.UUID `db:"challenge_id" json:"challenge_id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	Date        time.Time `db:"date"         json:"date"`
	Status      string    `db:"status"       json:"status"`
	Comment     *string   `db:"comment"      json:"comment"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}

type CheckInImage struct {
	ID        uuid.UUID `db:"id"          json:"id"`
	CheckInID uuid.UUID `db:"check_in_id" json:"check_in_id"`
	URL       string    `db:"url"         json:"url"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
}

// New simplified check-in (presence + optional comment)
type SimpleCheckIn struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	ChallengeID uuid.UUID `db:"challenge_id" json:"challenge_id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	Date        time.Time `db:"date"         json:"date"`
	Comment     string    `db:"comment"      json:"comment"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}

type Progress struct {
	CheckedInToday bool    `json:"checked_in_today"`
	IsWorkingDay   bool    `json:"is_working_day"`
	CurrentStreak  int     `json:"current_streak"`
	MaxStreak      int     `json:"max_streak"`
	DoneDays       int     `json:"done_days"`
	TotalDays      int     `json:"total_days"`
	AdherencePct   float64 `json:"adherence_pct"`
}