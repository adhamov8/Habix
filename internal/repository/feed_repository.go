package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type FeedRepository struct {
	db *sqlx.DB
}

func NewFeedRepository(db *sqlx.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) Insert(ctx context.Context, e *domain.FeedEvent) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO feed_events (id, challenge_id, user_id, type, reference_id, data)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID, e.ChallengeID, e.UserID, e.Type, e.ReferenceID, e.Data)
	return err
}

func (r *FeedRepository) ListByChallenge(ctx context.Context, challengeID uuid.UUID, limit, offset int) ([]domain.FeedEvent, error) {
	var list []domain.FeedEvent
	err := r.db.SelectContext(ctx, &list, `
		SELECT fe.*, u.name AS user_name,
			COALESCE((SELECT COUNT(*) FROM feed_comments fc WHERE fc.feed_event_id = fe.id), 0) AS comment_count
		FROM feed_events fe
		JOIN users u ON u.id = fe.user_id
		WHERE fe.challenge_id = $1
		ORDER BY fe.created_at DESC
		LIMIT $2 OFFSET $3`, challengeID, limit, offset)
	return list, err
}