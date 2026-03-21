package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO users (id, email, password_hash, name, timezone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Timezone,
	).Scan(&user.CreatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE email = $1", email); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	if err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET name = $1, bio = $2, timezone = $3 WHERE id = $4`,
		user.Name, user.Bio, user.Timezone, user.ID,
	)
	return err
}