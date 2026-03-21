package domain

import (
	"time"

	"github.com/google/uuid"
)

type FeedComment struct {
	ID          uuid.UUID `db:"id"            json:"id"`
	FeedEventID uuid.UUID `db:"feed_event_id" json:"feed_event_id"`
	UserID      uuid.UUID `db:"user_id"       json:"user_id"`
	Text        string    `db:"text"          json:"text"`
	CreatedAt   time.Time `db:"created_at"    json:"created_at"`
	// Joined
	UserName string `db:"user_name" json:"user_name"`
}
