package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type ParticipantRepository struct {
	db *sqlx.DB
}

func NewParticipantRepository(db *sqlx.DB) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

func (r *ParticipantRepository) Add(ctx context.Context, challengeID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO challenge_participants (challenge_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, challengeID, userID)
	return err
}

func (r *ParticipantRepository) Remove(ctx context.Context, challengeID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM challenge_participants WHERE challenge_id = $1 AND user_id = $2`,
		challengeID, userID)
	return err
}

func (r *ParticipantRepository) Exists(ctx context.Context, challengeID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS(SELECT 1 FROM challenge_participants WHERE challenge_id = $1 AND user_id = $2)`,
		challengeID, userID)
	return exists, err
}

func (r *ParticipantRepository) Count(ctx context.Context, challengeID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM challenge_participants WHERE challenge_id = $1`,
		challengeID)
	return count, err
}

func (r *ParticipantRepository) ListByChallenge(ctx context.Context, challengeID uuid.UUID) ([]domain.Participant, error) {
	var list []domain.Participant
	err := r.db.SelectContext(ctx, &list, `
		SELECT * FROM challenge_participants WHERE challenge_id = $1 ORDER BY joined_at`, challengeID)
	return list, err
}