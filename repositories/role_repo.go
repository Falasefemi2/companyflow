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

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{
		pool: pool,
	}
}

func (r *RoleRepository) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	INSERT INTO roles (
		company_id, name, description, is_system_role, permissions_cache
	)
	VALUES (
		$1,$2,$3,$4,$5
	)
	RETURNING id, company_id, name, description, is_system_role, permissions_cache, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		role.CompanyID,
		role.Name,
		role.Description,
		role.IsSystemRole,
		role.PermissionsCache,
	).Scan(
		&role.ID,
		&role.CompanyID,
		&role.Name,
		&role.Description,
		&role.IsSystemRole,
		&role.PermissionsCache,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RoleRepository) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, company_id, name, description, is_system_role, permissions_cache, created_at, updated_at
	FROM roles
	WHERE id = $1
	`

	var role models.Role

	err := r.pool.QueryRow(ctx, query, roleID).Scan(
		&role.ID,
		&role.CompanyID,
		&role.Name,
		&role.Description,
		&role.IsSystemRole,
		&role.PermissionsCache,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetRoleList returns all roles for a company with pagination and filtering
// Supports filtering by:
// - Search (name - case insensitive)
func (r *RoleRepository) GetRoleList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.RoleListRequest,
) (*utils.PaginatedResponse[*models.Role], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Build WHERE clause with filters
	where := "WHERE company_id = $1"
	args := []any{companyID}
	i := 2

	// Search in name if provided
	if listRequest.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", i)
		search := "%" + listRequest.Search + "%"
		args = append(args, search)
		i++
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM roles %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (listRequest.Page - 1) * listRequest.PageSize

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			id, company_id, name, description, is_system_role, permissions_cache, created_at, updated_at
		FROM roles
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, i, i+1)

	args = append(args, listRequest.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role

	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID, &role.CompanyID, &role.Name, &role.Description,
			&role.IsSystemRole, &role.PermissionsCache,
			&role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.Role]{
		Data:       roles,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (r *RoleRepository) UpdateRole(ctx context.Context, roleID uuid.UUID, role *models.Role) (*models.Role, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		UPDATE roles
		SET
			name = COALESCE(NULLIF($1, ''), name),
			description = COALESCE($2, description),
			permissions_cache = COALESCE($3::jsonb, permissions_cache),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, company_id, name, description, is_system_role, permissions_cache, created_at, updated_at
	`

	var updated models.Role
	err := r.pool.QueryRow(ctx, query,
		role.Name,
		role.Description,
		role.PermissionsCache,
		roleID,
	).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Name,
		&updated.Description,
		&updated.IsSystemRole,
		&updated.PermissionsCache,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *RoleRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Hard delete: remove from database
	result, err := r.pool.Exec(ctx, "DELETE FROM roles WHERE id = $1", roleID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("role not found")
	}

	return nil
}
