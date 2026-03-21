package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         string    `db:"name"          json:"name"`
	AvatarURL    *string   `db:"avatar_url"    json:"avatar_url"`
	Bio          *string   `db:"bio"           json:"bio"`
	Timezone     string    `db:"timezone"      json:"timezone"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}