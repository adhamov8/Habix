package repository

import (
	"context"
	"fmt"
	"strings"

	"tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ChallengeRepository struct {
	db *sqlx.DB
}

func NewChallengeRepository(db *sqlx.DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

const selectChallenge = `
	SELECT id, creator_id, category_id, title, description,
		starts_at, ends_at, working_days, max_skips,
		TO_CHAR(deadline_time, 'HH24:MI:SS') AS deadline_time,
		is_public, invite_token, status, created_at
	FROM challenges`

const selectChallengeC = `
	SELECT c.id, c.creator_id, c.category_id, c.title, c.description,
		c.starts_at, c.ends_at, c.working_days, c.max_skips,
		TO_CHAR(c.deadline_time, 'HH24:MI:SS') AS deadline_time,
		c.is_public, c.invite_token, c.status, c.created_at
	FROM challenges c`

func (r *ChallengeRepository) GetDeadlineTimeText(ctx context.Context, id uuid.UUID) (string, error) {
	var s string
	err := r.db.GetContext(ctx, &s,
		"SELECT TO_CHAR(deadline_time, 'HH24:MI:SS') FROM challenges WHERE id = $1", id)
	return s, err
}

func (r *ChallengeRepository) Create(ctx context.Context, c *domain.Challenge) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO challenges
			(id, creator_id, category_id, title, description,
			 starts_at, ends_at, working_days, max_skips, deadline_time,
			 is_public, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING invite_token, created_at`,
		c.ID, c.CreatorID, c.CategoryID, c.Title, c.Description,
		c.StartsAt, c.EndsAt, c.WorkingDays, c.MaxSkips, c.DeadlineTime,
		c.IsPublic, c.Status,
	).Scan(&c.InviteToken, &c.CreatedAt)
}

func (r *ChallengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Challenge, error) {
	var c domain.Challenge
	if err := r.db.GetContext(ctx, &c, selectChallenge+" WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ChallengeRepository) GetByInviteToken(ctx context.Context, token uuid.UUID) (*domain.Challenge, error) {
	var c domain.Challenge
	if err := r.db.GetContext(ctx, &c, selectChallenge+" WHERE invite_token = $1", token); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ChallengeRepository) Update(ctx context.Context, c *domain.Challenge) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE challenges SET
			category_id = $1, title = $2, description = $3,
			starts_at = $4, ends_at = $5, working_days = $6,
			max_skips = $7, deadline_time = $8, is_public = $9
		WHERE id = $10`,
		c.CategoryID, c.Title, c.Description,
		c.StartsAt, c.EndsAt, c.WorkingDays,
		c.MaxSkips, c.DeadlineTime, c.IsPublic,
		c.ID,
	)
	return err
}

func (r *ChallengeRepository) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE challenges SET status = $1 WHERE id = $2", status, id)
	return err
}

func (r *ChallengeRepository) ActivateUpcoming(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE challenges SET status = 'active'
		WHERE status = 'upcoming' AND starts_at <= CURRENT_DATE`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *ChallengeRepository) FinishExpired(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `
		UPDATE challenges SET status = 'finished'
		WHERE status = 'active' AND ends_at < CURRENT_DATE
		RETURNING id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *ChallengeRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM challenges WHERE status = 'active'")
	return count, err
}

func (r *ChallengeRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]domain.Challenge, error) {
	var list []domain.Challenge
	err := r.db.SelectContext(ctx, &list, `
		SELECT DISTINCT c.id, c.creator_id, c.category_id, c.title, c.description,
			c.starts_at, c.ends_at, c.working_days, c.max_skips,
			TO_CHAR(c.deadline_time, 'HH24:MI:SS') AS deadline_time,
			c.is_public, c.invite_token, c.status, c.created_at
		FROM challenges c
		LEFT JOIN challenge_participants cp ON cp.challenge_id = c.id
		WHERE c.creator_id = $1 OR cp.user_id = $1
		ORDER BY c.created_at DESC`, userID)
	return list, err
}

func (r *ChallengeRepository) ListPublic(ctx context.Context, categoryID *int, search string, limit, offset int) ([]domain.Challenge, error) {
	var (
		clauses []string
		args    []interface{}
		argIdx  int
	)

	clauses = append(clauses, "c.is_public = true")

	if categoryID != nil {
		argIdx++
		clauses = append(clauses, fmt.Sprintf("c.category_id = $%d", argIdx))
		args = append(args, *categoryID)
	}
	if search != "" {
		argIdx++
		clauses = append(clauses, fmt.Sprintf("c.title ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
	}

	argIdx++
	limitArg := argIdx
	argIdx++
	offsetArg := argIdx
	args = append(args, limit, offset)

	query := fmt.Sprintf(`%s
		WHERE %s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d`,
		selectChallengeC, strings.Join(clauses, " AND "), limitArg, offsetArg)

	var list []domain.Challenge
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}
