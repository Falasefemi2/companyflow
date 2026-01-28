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

type EmployeeRepository struct {
	pool *pgxpool.Pool
}

func NewEmployeeRepository(pool *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{pool: pool}
}

func (e *EmployeeRepository) CreateEmployee(ctx context.Context, employee *models.Employee) (*models.Employee, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		INSERT INTO employees (
			company_id, email, password_hash, phone, first_name, last_name,
			employee_code, department_id, designation_id, level_id, manager_id,
			role_id, status, employment_type, hire_date, date_of_birth, gender,
			address, emergency_contact_name, emergency_contact_phone, profile_image_url
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21
		)
		RETURNING id, company_id, email, password_hash, phone, first_name, last_name,
				  employee_code, department_id, designation_id, level_id, manager_id,
				  role_id, status, employment_type, hire_date, termination_date,
				  date_of_birth, gender, address, emergency_contact_name,
				  emergency_contact_phone, profile_image_url, last_login_at,
				  created_at, updated_at
	`

	err := e.pool.QueryRow(ctx, query,
		employee.CompanyID,
		employee.Email,
		employee.PasswordHash,
		employee.Phone,
		employee.FirstName,
		employee.LastName,
		employee.EmployeeCode,
		employee.DepartmentID,
		employee.DesignationID,
		employee.LevelID,
		employee.ManagerID,
		employee.RoleID,
		employee.Status,
		employee.EmploymentType,
		employee.HireDate,
		employee.DateOfBirth,
		employee.Gender,
		employee.Address,
		employee.EmergencyContactName,
		employee.EmergencyContactPhone,
		employee.ProfileImageURL,
	).Scan(
		&employee.ID,
		&employee.CompanyID,
		&employee.Email,
		&employee.PasswordHash,
		&employee.Phone,
		&employee.FirstName,
		&employee.LastName,
		&employee.EmployeeCode,
		&employee.DepartmentID,
		&employee.DesignationID,
		&employee.LevelID,
		&employee.ManagerID,
		&employee.RoleID,
		&employee.Status,
		&employee.EmploymentType,
		&employee.HireDate,
		&employee.TerminationDate,
		&employee.DateOfBirth,
		&employee.Gender,
		&employee.Address,
		&employee.EmergencyContactName,
		&employee.EmergencyContactPhone,
		&employee.ProfileImageURL,
		&employee.LastLoginAt,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return employee, nil
}

func (e *EmployeeRepository) GetEmployeeByID(ctx context.Context, employeeID uuid.UUID) (*models.Employee, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		SELECT
			id, company_id, email, password_hash, phone, first_name, last_name,
			employee_code, department_id, designation_id, level_id, manager_id,
			role_id, status, employment_type, hire_date, termination_date,
			date_of_birth, gender, address, emergency_contact_name,
			emergency_contact_phone, profile_image_url, last_login_at,
			created_at, updated_at
		FROM employees
		WHERE id = $1
	`

	var employee models.Employee

	err := e.pool.QueryRow(ctx, query, employeeID).Scan(
		&employee.ID,
		&employee.CompanyID,
		&employee.Email,
		&employee.PasswordHash,
		&employee.Phone,
		&employee.FirstName,
		&employee.LastName,
		&employee.EmployeeCode,
		&employee.DepartmentID,
		&employee.DesignationID,
		&employee.LevelID,
		&employee.ManagerID,
		&employee.RoleID,
		&employee.Status,
		&employee.EmploymentType,
		&employee.HireDate,
		&employee.TerminationDate,
		&employee.DateOfBirth,
		&employee.Gender,
		&employee.Address,
		&employee.EmergencyContactName,
		&employee.EmergencyContactPhone,
		&employee.ProfileImageURL,
		&employee.LastLoginAt,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &employee, nil
}

func (e *EmployeeRepository) GetEmployeeList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.EmployeeListRequest,
) (*utils.PaginatedResponse[*models.Employee], error) {

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	where := "WHERE company_id = $1"
	args := []any{companyID}
	i := 2

	if listRequest.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, listRequest.Status)
		i++
	}

	if listRequest.Search != "" {
		where += fmt.Sprintf(
			" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)",
			i, i, i,
		)
		search := "%" + listRequest.Search + "%"
		args = append(args, search, search, search)
		i += 3
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees %s", where)
	if err := e.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (listRequest.Page - 1) * listRequest.PageSize

	query := fmt.Sprintf(`
		SELECT
			id, company_id, email, password_hash, phone, first_name, last_name,
			employee_code, department_id, designation_id, level_id, manager_id,
			role_id, status, employment_type, hire_date, termination_date,
			date_of_birth, gender, address, emergency_contact_name,
			emergency_contact_phone, profile_image_url, last_login_at,
			created_at, updated_at
		FROM employees
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, i, i+1)

	args = append(args, listRequest.PageSize, offset)

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*models.Employee

	for rows.Next() {
		var emp models.Employee
		if err := rows.Scan(
			&emp.ID, &emp.CompanyID, &emp.Email, &emp.PasswordHash,
			&emp.Phone, &emp.FirstName, &emp.LastName, &emp.EmployeeCode,
			&emp.DepartmentID, &emp.DesignationID, &emp.LevelID, &emp.ManagerID,
			&emp.RoleID, &emp.Status, &emp.EmploymentType, &emp.HireDate,
			&emp.TerminationDate, &emp.DateOfBirth, &emp.Gender, &emp.Address,
			&emp.EmergencyContactName, &emp.EmergencyContactPhone,
			&emp.ProfileImageURL, &emp.LastLoginAt, &emp.CreatedAt, &emp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		employees = append(employees, &emp)
	}

	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.Employee]{
		Data:       employees,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (e *EmployeeRepository) DeleteEmployee(ctx context.Context, employeeID string, hardDelete bool) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if hardDelete {
		result, err := e.pool.Exec(ctx, "DELETE FROM employees WHERE id = $1", employeeID)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return errors.New("employee not found")
		}
		return nil
	}

	result, err := e.pool.Exec(
		ctx,
		"UPDATE employees SET status = 'inactive', updated_at = CURRENT_TIMESTAMP WHERE id = $1",
		employeeID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("employee not found")
	}

	return nil
}
