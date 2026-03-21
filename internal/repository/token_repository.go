package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type TokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token *RefreshToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt,
	)
	return err
}

func (r *TokenRepository) GetByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	var t RefreshToken
	err := r.db.GetContext(ctx, &t,
		"SELECT * FROM refresh_tokens WHERE token_hash = $1 AND expires_at > NOW()", hash)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) Delete(ctx context.Context, hash string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token_hash = $1", hash)
	return err
}