package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/falasefemi2/companyflowlow/models"
)

type LevelRepository struct {
	pool *pgxpool.Pool
}

func NewLevelRepository(pool *pgxpool.Pool) *LevelRepository {
	return &LevelRepository{
		pool: pool,
	}
}

func (l *LevelRepository) CreateLevel(ctx context.Context, level *models.Level) (*models.Level, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	INSERT INTO levels (
		company_id, name, hierarchy_level, min_salary, max_salary,description
	)
	VALUES (
		$1,$2,$3,$4,$5,$6
	)
	RETURNING id, 
		company_id, name, hierarchy_level, min_salary, max_salary,description, created_at, updated_at
	`

	err := l.pool.QueryRow(ctx, query,
		level.CompanyID,
		level.Name,
		level.HierarchyLevel,
		level.MinSalary,
		level.MaxSalary,
		level.Description,
	).Scan(
		&level.ID,
		&level.CompanyID,
		&level.Name,
		&level.HierarchyLevel,
		&level.MinSalary,
		&level.MaxSalary,
		&level.Description,
		&level.CreatedAt,
		&level.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return level, nil
}
