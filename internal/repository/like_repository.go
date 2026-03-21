package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type LikeRepository struct {
	db *sqlx.DB
}

func NewLikeRepository(db *sqlx.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

// Toggle inserts or deletes a like. Returns true if liked, false if unliked.
func (r *LikeRepository) Toggle(ctx context.Context, checkInID, userID uuid.UUID) (bool, error) {
	var deleted bool
	err := r.db.QueryRowContext(ctx, `
		WITH del AS (
			DELETE FROM check_in_likes
			WHERE check_in_id = $1 AND user_id = $2
			RETURNING true
		)
		SELECT EXISTS (SELECT 1 FROM del)`,
		checkInID, userID,
	).Scan(&deleted)
	if err != nil {
		return false, err
	}
	if deleted {
		return false, nil
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO check_in_likes (check_in_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, checkInID, userID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *LikeRepository) Count(ctx context.Context, checkInID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM check_in_likes WHERE check_in_id = $1", checkInID)
	return count, err
}