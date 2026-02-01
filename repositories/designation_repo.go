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

type DesignationRepository struct {
	pool *pgxpool.Pool
}

func NewDesignationRepository(pool *pgxpool.Pool) *DesignationRepository {
	return &DesignationRepository{
		pool: pool,
	}
}

func (d *DesignationRepository) CreateDesignation(ctx context.Context, designation *models.Designation) (*models.Designation, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		INSERT INTO designations (
			company_id, name, description, level_id, department_id, status
		)
		VALUES (
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id, company_id, name, description, level_id, department_id, status, created_at, updated_at
	`

	err := d.pool.QueryRow(ctx, query,
		designation.CompanyID,
		designation.Name,
		designation.Description,
		designation.LevelID,
		designation.DepartmentID,
		designation.Status,
	).Scan(
		&designation.ID,
		&designation.CompanyID,
		&designation.Name,
		&designation.Description,
		&designation.LevelID,
		&designation.DepartmentID,
		&designation.Status,
		&designation.CreatedAt,
		&designation.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return designation, nil
}

func (d *DesignationRepository) GetDesignationByID(ctx context.Context, designationID uuid.UUID) (*models.Designation, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		SELECT id, company_id, name, description, level_id, department_id, status, created_at, updated_at
		FROM designations
		WHERE id = $1
	`

	var designation models.Designation

	err := d.pool.QueryRow(ctx, query, designationID).Scan(
		&designation.ID,
		&designation.CompanyID,
		&designation.Name,
		&designation.Description,
		&designation.LevelID,
		&designation.DepartmentID,
		&designation.Status,
		&designation.CreatedAt,
		&designation.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &designation, nil
}

func (d *DesignationRepository) GetDesignationList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.DesignationListRequest,
) (*utils.PaginatedResponse[*models.Designation], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	where := "WHERE company_id = $1"
	args := []any{companyID}
	i := 2

	// Filter by status if provided
	if listRequest.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, listRequest.Status)
		i++
	}

	// Filter by level_id if provided
	if listRequest.LevelID != "" {
		where += fmt.Sprintf(" AND level_id = $%d::uuid", i)
		args = append(args, listRequest.LevelID)
		i++
	}

	// Filter by department_id if provided
	if listRequest.DepartmentID != "" {
		where += fmt.Sprintf(" AND department_id = $%d::uuid", i)
		args = append(args, listRequest.DepartmentID)
		i++
	}

	// Search in name if provided
	if listRequest.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", i)
		search := "%" + listRequest.Search + "%"
		args = append(args, search)
		i++
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM designations %s", where)
	if err := d.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (listRequest.Page - 1) * listRequest.PageSize

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			id, company_id, name, description, level_id, department_id, status, created_at, updated_at
		FROM designations
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, i, i+1)

	args = append(args, listRequest.PageSize, offset)

	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var designations []*models.Designation

	for rows.Next() {
		var des models.Designation
		if err := rows.Scan(
			&des.ID, &des.CompanyID, &des.Name, &des.Description,
			&des.LevelID, &des.DepartmentID, &des.Status,
			&des.CreatedAt, &des.UpdatedAt,
		); err != nil {
			return nil, err
		}
		designations = append(designations, &des)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.Designation]{
		Data:       designations,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (d *DesignationRepository) UpdateDesignation(ctx context.Context, designationID uuid.UUID, designation *models.Designation) (*models.Designation, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		UPDATE designations
		SET
			name = COALESCE(NULLIF($1, ''), name),
			description = COALESCE(NULLIF($2, ''), description),
			level_id = CASE WHEN $3::uuid IS DISTINCT FROM NULL THEN $3::uuid ELSE level_id END,
			department_id = CASE WHEN $4::uuid IS DISTINCT FROM NULL THEN $4::uuid ELSE department_id END,
			status = COALESCE(NULLIF($5, ''), status),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING id, company_id, name, description, level_id, department_id, status, created_at, updated_at
	`

	var updated models.Designation
	err := d.pool.QueryRow(ctx, query,
		designation.Name,
		designation.Description,
		designation.LevelID,
		designation.DepartmentID,
		designation.Status,
		designationID,
	).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Name,
		&updated.Description,
		&updated.LevelID,
		&updated.DepartmentID,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (d *DesignationRepository) DeleteDesignation(ctx context.Context, designationID uuid.UUID) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	result, err := d.pool.Exec(ctx, "DELETE FROM designations WHERE id = $1", designationID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("designation not found")
	}

	return nil
}
