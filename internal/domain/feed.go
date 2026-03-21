package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FeedEvent struct {
	ID           uuid.UUID        `db:"id"            json:"id"`
	ChallengeID  uuid.UUID        `db:"challenge_id"  json:"challenge_id"`
	UserID       uuid.UUID        `db:"user_id"       json:"user_id"`
	Type         string           `db:"type"          json:"type"`
	ReferenceID  *uuid.UUID       `db:"reference_id"  json:"reference_id"`
	Data         *json.RawMessage `db:"data"          json:"data"          swaggertype:"object"`
	CreatedAt    time.Time        `db:"created_at"    json:"created_at"`
	// Joined from users table
	UserName     string           `db:"user_name"     json:"user_name"`
	CommentCount int              `db:"comment_count" json:"comment_count"`
}