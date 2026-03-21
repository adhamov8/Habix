package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type FeedCommentRepository struct {
	db *sqlx.DB
}

func NewFeedCommentRepository(db *sqlx.DB) *FeedCommentRepository {
	return &FeedCommentRepository{db: db}
}

func (r *FeedCommentRepository) Create(ctx context.Context, c *domain.FeedComment) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO feed_comments (id, feed_event_id, user_id, text)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		c.ID, c.FeedEventID, c.UserID, c.Text,
	).Scan(&c.CreatedAt)
}

func (r *FeedCommentRepository) ListForEvent(ctx context.Context, feedEventID uuid.UUID) ([]domain.FeedComment, error) {
	var list []domain.FeedComment
	err := r.db.SelectContext(ctx, &list, `
		SELECT fc.id, fc.feed_event_id, fc.user_id, fc.text, fc.created_at,
			u.name AS user_name
		FROM feed_comments fc
		JOIN users u ON u.id = fc.user_id
		WHERE fc.feed_event_id = $1
		ORDER BY fc.created_at ASC`, feedEventID)
	return list, err
}

func (r *FeedCommentRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM feed_comments WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}
