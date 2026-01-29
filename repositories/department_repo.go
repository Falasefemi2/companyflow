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

type DepartmentRepository struct {
	pool *pgxpool.Pool
}

func NewDepartmentRepository(pool *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{
		pool: pool,
	}
}

func (d *DepartmentRepository) CreateDepartment(ctx context.Context, department *models.Department) (*models.Department, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		INSERT INTO departments (
			company_id, name, code, description, parent_department_id, cost_center, status
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7
		)
		RETURNING id, company_id, name, code, description, parent_department_id, cost_center, status, created_at, updated_at
	`

	err := d.pool.QueryRow(ctx, query,
		department.CompanyID,
		department.Name,
		department.Code,
		department.Description,
		department.ParentDepartmentID,
		department.CostCenter,
		department.Status,
	).Scan(
		&department.ID,
		&department.CompanyID,
		&department.Name,
		&department.Code,
		&department.Description,
		&department.ParentDepartmentID,
		&department.CostCenter,
		&department.Status,
		&department.CreatedAt,
		&department.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return department, nil
}

func (d *DepartmentRepository) GetDepartmentByID(ctx context.Context, departmentID uuid.UUID) (*models.Department, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		SELECT id, company_id, name, code, description, parent_department_id, cost_center, status, created_at, updated_at
		FROM departments
		WHERE id = $1
	`

	var department models.Department

	err := d.pool.QueryRow(ctx, query, departmentID).Scan(
		&department.ID,
		&department.CompanyID,
		&department.Name,
		&department.Code,
		&department.Description,
		&department.ParentDepartmentID,
		&department.CostCenter,
		&department.Status,
		&department.CreatedAt,
		&department.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &department, nil
}

// GetDepartmentList returns all departments for a company with pagination and filtering
// Supports filtering by:
// - Status (active, inactive)
// - Search (name or code - case insensitive)
func (d *DepartmentRepository) GetDepartmentList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.DepartmentListRequest,
) (*utils.PaginatedResponse[*models.Department], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Build WHERE clause with filters
	where := "WHERE company_id = $1"
	args := []any{companyID}
	i := 2

	// Filter by status if provided
	if listRequest.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, listRequest.Status)
		i++
	}

	// Search in name or code if provided
	if listRequest.Search != "" {
		where += fmt.Sprintf(
			" AND (name ILIKE $%d OR code ILIKE $%d)",
			i, i,
		)
		search := "%" + listRequest.Search + "%"
		args = append(args, search, search)
		i += 2
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM departments %s", where)
	if err := d.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (listRequest.Page - 1) * listRequest.PageSize

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			id, company_id, name, code, description, parent_department_id, cost_center, status, created_at, updated_at
		FROM departments
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

	var departments []*models.Department

	for rows.Next() {
		var dept models.Department
		if err := rows.Scan(
			&dept.ID, &dept.CompanyID, &dept.Name, &dept.Code,
			&dept.Description, &dept.ParentDepartmentID, &dept.CostCenter,
			&dept.Status, &dept.CreatedAt, &dept.UpdatedAt,
		); err != nil {
			return nil, err
		}
		departments = append(departments, &dept)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.Department]{
		Data:       departments,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (d *DepartmentRepository) UpdateDepartment(ctx context.Context, departmentID uuid.UUID, department *models.Department) (*models.Department, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		UPDATE departments
		SET
			name = COALESCE(NULLIF($1, ''), name),
			code = COALESCE(NULLIF($2, ''), code),
			description = COALESCE(NULLIF($3, ''), description),
			parent_department_id = CASE WHEN $4::uuid IS DISTINCT FROM NULL THEN $4::uuid ELSE parent_department_id END,
			cost_center = COALESCE(NULLIF($5, ''), cost_center),
			status = COALESCE(NULLIF($6, ''), status),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING id, company_id, name, code, description, parent_department_id, cost_center, status, created_at, updated_at
	`

	var updated models.Department
	err := d.pool.QueryRow(ctx, query,
		department.Name,
		department.Code,
		department.Description,
		department.ParentDepartmentID,
		department.CostCenter,
		department.Status,
		departmentID,
	).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Name,
		&updated.Code,
		&updated.Description,
		&updated.ParentDepartmentID,
		&updated.CostCenter,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (d *DepartmentRepository) DeleteDepartment(ctx context.Context, departmentID uuid.UUID, softDelete bool) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if softDelete {
		// Soft delete: mark as inactive
		result, err := d.pool.Exec(
			ctx,
			"UPDATE departments SET status = 'inactive', updated_at = CURRENT_TIMESTAMP WHERE id = $1",
			departmentID,
		)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return errors.New("department not found")
		}
		return nil
	}

	// Hard delete: remove from database
	result, err := d.pool.Exec(ctx, "DELETE FROM departments WHERE id = $1", departmentID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("department not found")
	}

	return nil
}
