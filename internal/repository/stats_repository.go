package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type StatsRepository struct {
	db *sqlx.DB
}

func NewStatsRepository(db *sqlx.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

type participantCounts struct {
	UserID   uuid.UUID `db:"user_id"`
	UserName string    `db:"user_name"`
	DoneDays int       `db:"done_days"`
}

// GetParticipantCounts returns done days per participant using new checkins table.
func (r *StatsRepository) GetParticipantCounts(ctx context.Context, challengeID uuid.UUID) ([]participantCounts, error) {
	var list []participantCounts
	err := r.db.SelectContext(ctx, &list, `
		SELECT
			cp.user_id,
			u.name AS user_name,
			COUNT(ci.id) AS done_days
		FROM challenge_participants cp
		JOIN users u ON u.id = cp.user_id
		LEFT JOIN checkins ci ON ci.challenge_id = cp.challenge_id AND ci.user_id = cp.user_id
		WHERE cp.challenge_id = $1
		GROUP BY cp.user_id, u.name`, challengeID)
	return list, err
}

type userCheckIn struct {
	UserID uuid.UUID `db:"user_id"`
	Date   time.Time `db:"date"`
}

// GetCheckInsForChallenge returns all check-ins ordered by user and date.
func (r *StatsRepository) GetCheckInsForChallenge(ctx context.Context, challengeID uuid.UUID) ([]userCheckIn, error) {
	var list []userCheckIn
	err := r.db.SelectContext(ctx, &list, `
		SELECT user_id, date FROM checkins
		WHERE challenge_id = $1
		ORDER BY user_id, date`, challengeID)
	return list, err
}

type userChallengeStats struct {
	ChallengeID uuid.UUID `db:"challenge_id"`
	Status      string    `db:"status"`
	DoneDays    int       `db:"done_days"`
	TotalDays   int       `db:"total_days"`
}

// GetUserChallengeStats returns per-challenge stats for a user.
func (r *StatsRepository) GetUserChallengeStats(ctx context.Context, userID uuid.UUID) ([]userChallengeStats, error) {
	var list []userChallengeStats
	err := r.db.SelectContext(ctx, &list, `
		SELECT DISTINCT
			c.id AS challenge_id,
			c.status,
			COALESCE((SELECT COUNT(*) FROM checkins ci
				WHERE ci.challenge_id = c.id AND ci.user_id = $1), 0) AS done_days,
			COALESCE((SELECT COUNT(*) FROM checkins ci
				WHERE ci.challenge_id = c.id AND ci.user_id = $1), 0) AS total_days
		FROM challenges c
		LEFT JOIN challenge_participants cp ON cp.challenge_id = c.id AND cp.user_id = $1
		WHERE c.creator_id = $1 OR cp.user_id = $1`, userID)
	return list, err
}

// GetUserAllCheckIns returns all check-ins for a user across all challenges.
func (r *StatsRepository) GetUserAllCheckIns(ctx context.Context, userID uuid.UUID) ([]userCheckIn, error) {
	var list []userCheckIn
	err := r.db.SelectContext(ctx, &list, `
		SELECT user_id, date FROM checkins
		WHERE user_id = $1
		ORDER BY date`, userID)
	return list, err
}

// GetParticipationByDay returns daily participation counts.
func (r *StatsRepository) GetParticipationByDay(ctx context.Context, challengeID uuid.UUID) ([]domain.DayParticipation, error) {
	var list []domain.DayParticipation
	err := r.db.SelectContext(ctx, &list, `
		WITH participant_count AS (
			SELECT COUNT(*) AS cnt FROM challenge_participants WHERE challenge_id = $1
		)
		SELECT
			ci.date::text AS date,
			COUNT(ci.id) AS checked_in,
			(SELECT cnt FROM participant_count) AS total_users
		FROM checkins ci
		WHERE ci.challenge_id = $1
		GROUP BY ci.date
		ORDER BY ci.date`, challengeID)
	return list, err
}