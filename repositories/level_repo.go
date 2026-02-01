package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/utils"
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

func (l *LevelRepository) GetLevelByID(ctx context.Context, levelID uuid.UUID) (*models.Level, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, company_id, name, hierarchy_level, min_salary, max_salary,description, created_at, updated_at
	FROM levels
	WHERE id = $1 
`
	var level models.Level
	err := l.pool.QueryRow(ctx, query, levelID).Scan(
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
	return &level, nil
}

func (l *LevelRepository) GetLevelList(ctx context.Context, companyID uuid.UUID, listRequest *dto.LevelListRequest) (*utils.PaginatedResponse[*models.Level], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	where := "WHERE company_id = $1"
	args := []any{companyID}
	i := 2

	if listRequest.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", i)
		search := "%" + listRequest.Search + "%"
		args = append(args, search)
		i++
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM levels %s", where)
	if err := l.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (listRequest.Page - 1) * listRequest.PageSize

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			id, company_id, name, hierarchy_level, min_salary, max_salary, description, created_at, updated_at
		FROM levels
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, i, i+1)

	args = append(args, listRequest.PageSize, offset)

	rows, err := l.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []*models.Level

	for rows.Next() {
		var level models.Level
		if err := rows.Scan(
			&level.ID, &level.CompanyID, &level.Name, &level.HierarchyLevel,
			&level.MinSalary, &level.MaxSalary, &level.Description,
			&level.CreatedAt, &level.UpdatedAt,
		); err != nil {
			return nil, err
		}
		levels = append(levels, &level)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.Level]{
		Data:       levels,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (l *LevelRepository) UpdateLevel(ctx context.Context, levelID uuid.UUID, level *models.Level) (*models.Level, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		UPDATE levels
		SET
			name = COALESCE(NULLIF($1, ''), name),
			hierarchy_level = CASE WHEN $2 > 0 THEN $2 ELSE hierarchy_level END,
			min_salary = COALESCE($3::numeric, min_salary),
			max_salary = COALESCE($4::numeric, max_salary),
			description = COALESCE(NULLIF($5, ''), description),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING id, company_id, name, hierarchy_level, min_salary, max_salary, description, created_at, updated_at
	`

	var updated models.Level
	err := l.pool.QueryRow(ctx, query,
		level.Name,
		level.HierarchyLevel,
		level.MinSalary,
		level.MaxSalary,
		level.Description,
		levelID,
	).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Name,
		&updated.HierarchyLevel,
		&updated.MinSalary,
		&updated.MaxSalary,
		&updated.Description,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (l *LevelRepository) DeleteLevel(ctx context.Context, levelID uuid.UUID) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	result, err := l.pool.Exec(ctx, "DELETE FROM levels WHERE id = $1", levelID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("level not found")
	}

	return nil
}
