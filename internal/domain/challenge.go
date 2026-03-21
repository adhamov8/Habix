package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Category struct {
	ID   int    `db:"id"   json:"id"`
	Name string `db:"name" json:"name"`
}

type Challenge struct {
	ID           uuid.UUID     `db:"id"            json:"id"`
	CreatorID    uuid.UUID     `db:"creator_id"    json:"creator_id"`
	CategoryID   int           `db:"category_id"   json:"category_id"`
	Title        string        `db:"title"         json:"title"`
	Description  *string       `db:"description"   json:"description"`
	StartsAt     time.Time     `db:"starts_at"     json:"starts_at"`
	EndsAt       time.Time     `db:"ends_at"       json:"ends_at"`
	WorkingDays  pq.Int64Array `db:"working_days"  json:"working_days"  swaggertype:"array,integer"`
	MaxSkips     int           `db:"max_skips"     json:"max_skips"`
	DeadlineTime string        `db:"deadline_time" json:"deadline_time"`
	IsPublic     bool          `db:"is_public"     json:"is_public"`
	InviteToken  uuid.UUID     `db:"invite_token"  json:"invite_token,omitempty"`
	Status       string        `db:"status"        json:"status"`
	CreatedAt    time.Time     `db:"created_at"    json:"created_at"`
}

type Participant struct {
	ChallengeID uuid.UUID `db:"challenge_id" json:"challenge_id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	JoinedAt    time.Time `db:"joined_at"    json:"joined_at"`
}