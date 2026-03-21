package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"tracker/internal/domain"
)

type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	var cats []domain.Category
	if err := r.db.SelectContext(ctx, &cats, "SELECT id, name FROM categories ORDER BY id"); err != nil {
		return nil, err
	}
	return cats, nil
}