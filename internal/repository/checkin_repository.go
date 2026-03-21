package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type CheckInRepository struct {
	db *sqlx.DB
}

func NewCheckInRepository(db *sqlx.DB) *CheckInRepository {
	return &CheckInRepository{db: db}
}

// Create inserts a new check-in. Returns the created record.
func (r *CheckInRepository) Create(ctx context.Context, ci *domain.SimpleCheckIn) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO checkins (id, challenge_id, user_id, date, comment)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`,
		ci.ID, ci.ChallengeID, ci.UserID, ci.Date, ci.Comment,
	).Scan(&ci.CreatedAt)
}

// Delete removes a check-in by challenge, user, and date.
func (r *CheckInRepository) Delete(ctx context.Context, challengeID, userID uuid.UUID, date time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM checkins
		WHERE challenge_id = $1 AND user_id = $2 AND date = $3`,
		challengeID, userID, date)
	return err
}

// ExistsForDate checks whether a check-in exists for the given (challenge, user, date).
func (r *CheckInRepository) ExistsForDate(ctx context.Context, challengeID, userID uuid.UUID, date time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM checkins
			WHERE challenge_id = $1 AND user_id = $2 AND date = $3
		)`, challengeID, userID, date).Scan(&exists)
	return exists, err
}

// ListForUser returns all check-in dates for a user in a challenge.
func (r *CheckInRepository) ListForUser(ctx context.Context, challengeID, userID uuid.UUID) ([]domain.SimpleCheckIn, error) {
	var list []domain.SimpleCheckIn
	err := r.db.SelectContext(ctx, &list, `
		SELECT id, challenge_id, user_id, date, comment, created_at
		FROM checkins
		WHERE challenge_id = $1 AND user_id = $2
		ORDER BY date DESC`, challengeID, userID)
	return list, err
}

// ListForChallenge returns all check-ins in a challenge.
func (r *CheckInRepository) ListForChallenge(ctx context.Context, challengeID uuid.UUID) ([]domain.SimpleCheckIn, error) {
	var list []domain.SimpleCheckIn
	err := r.db.SelectContext(ctx, &list, `
		SELECT id, challenge_id, user_id, date, comment, created_at
		FROM checkins
		WHERE challenge_id = $1
		ORDER BY user_id, date`, challengeID)
	return list, err
}

// CountByUser returns the number of check-ins per user in a challenge.
func (r *CheckInRepository) CountByUser(ctx context.Context, challengeID uuid.UUID) (map[uuid.UUID]int, error) {
	type row struct {
		UserID uuid.UUID `db:"user_id"`
		Count  int       `db:"cnt"`
	}
	var rows []row
	err := r.db.SelectContext(ctx, &rows, `
		SELECT user_id, COUNT(*) AS cnt
		FROM checkins
		WHERE challenge_id = $1
		GROUP BY user_id`, challengeID)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		m[r.UserID] = r.Count
	}
	return m, nil
}

// GetByID returns a check-in by its ID (used for feed reference validation).
func (r *CheckInRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SimpleCheckIn, error) {
	var ci domain.SimpleCheckIn
	if err := r.db.GetContext(ctx, &ci, `
		SELECT id, challenge_id, user_id, date, comment, created_at
		FROM checkins WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &ci, nil
}