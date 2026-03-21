package domain

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID `db:"id"          json:"id"`
	CheckInID uuid.UUID `db:"check_in_id" json:"check_in_id"`
	UserID    uuid.UUID `db:"user_id"     json:"user_id"`
	Text      string    `db:"text"        json:"text"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
	// Joined from users table
	UserName  string    `db:"user_name"   json:"user_name"`
}

type Like struct {
	CheckInID uuid.UUID `db:"check_in_id" json:"check_in_id"`
	UserID    uuid.UUID `db:"user_id"     json:"user_id"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
}