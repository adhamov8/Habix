package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type BadgeRepository struct {
	db *sqlx.DB
}

func NewBadgeRepository(db *sqlx.DB) *BadgeRepository {
	return &BadgeRepository{db: db}
}

func (r *BadgeRepository) ListDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error) {
	var list []domain.BadgeDefinition
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM badge_definitions ORDER BY id`)
	return list, err
}

func (r *BadgeRepository) GetDefinitionByCode(ctx context.Context, code string) (*domain.BadgeDefinition, error) {
	var bd domain.BadgeDefinition
	err := r.db.GetContext(ctx, &bd, `SELECT * FROM badge_definitions WHERE code = $1`, code)
	if err != nil {
		return nil, err
	}
	return &bd, nil
}

// Award gives a badge to a user. Uses ON CONFLICT DO NOTHING to prevent duplicates.
// Returns true if the badge was actually awarded (not a duplicate).
func (r *BadgeRepository) Award(ctx context.Context, userID uuid.UUID, badgeCode string, challengeID *uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO user_badges (user_id, badge_id, challenge_id)
		SELECT $1, bd.id, $3
		FROM badge_definitions bd
		WHERE bd.code = $2
		ON CONFLICT DO NOTHING`,
		userID, badgeCode, challengeID)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func (r *BadgeRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	var list []domain.UserBadge
	err := r.db.SelectContext(ctx, &list, `
		SELECT ub.*, bd.code, bd.title, bd.description, bd.icon
		FROM user_badges ub
		JOIN badge_definitions bd ON bd.id = ub.badge_id
		WHERE ub.user_id = $1
		ORDER BY ub.earned_at DESC`, userID)
	return list, err
}

func (r *BadgeRepository) ListRecent(ctx context.Context, limit int) ([]domain.UserBadge, error) {
	var list []domain.UserBadge
	err := r.db.SelectContext(ctx, &list, `
		SELECT ub.*, bd.code, bd.title, bd.description, bd.icon
		FROM user_badges ub
		JOIN badge_definitions bd ON bd.id = ub.badge_id
		ORDER BY ub.earned_at DESC
		LIMIT $1`, limit)
	return list, err
}

// CountUserChallenges returns the number of challenges the user participates in.
func (r *BadgeRepository) CountUserChallenges(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM challenge_participants WHERE user_id = $1`, userID)
	return count, err
}