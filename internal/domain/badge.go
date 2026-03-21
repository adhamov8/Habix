package domain

import (
	"time"

	"github.com/google/uuid"
)

type BadgeDefinition struct {
	ID          int    `db:"id"          json:"id"`
	Code        string `db:"code"        json:"code"`
	Title       string `db:"title"       json:"title"`
	Description string `db:"description" json:"description"`
	Icon        string `db:"icon"        json:"icon"`
}

type UserBadge struct {
	ID           uuid.UUID  `db:"id"           json:"id"`
	UserID       uuid.UUID  `db:"user_id"      json:"user_id"`
	BadgeID      int        `db:"badge_id"     json:"badge_id"`
	ChallengeID  *uuid.UUID `db:"challenge_id" json:"challenge_id"`
	EarnedAt     time.Time  `db:"earned_at"    json:"earned_at"`
	// Joined fields
	Code         string     `db:"code"         json:"code"`
	Title        string     `db:"title"        json:"title"`
	Description  string     `db:"description"  json:"description"`
	Icon         string     `db:"icon"         json:"icon"`
}