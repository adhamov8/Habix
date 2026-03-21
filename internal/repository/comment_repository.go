package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type CommentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, c *domain.Comment) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO check_in_comments (id, check_in_id, user_id, text)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		c.ID, c.CheckInID, c.UserID, c.Text,
	).Scan(&c.CreatedAt)
}

func (r *CommentRepository) ListByCheckIn(ctx context.Context, checkInID uuid.UUID) ([]domain.Comment, error) {
	var list []domain.Comment
	err := r.db.SelectContext(ctx, &list, `
		SELECT c.*, u.name AS user_name
		FROM check_in_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.check_in_id = $1
		ORDER BY c.created_at ASC`, checkInID)
	return list, err
}