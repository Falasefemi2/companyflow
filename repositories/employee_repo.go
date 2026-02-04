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

func (e *EmployeeRepository) GetRoleByID(ctx context.Context, roleID uuid.UUID) (string, *uuid.UUID, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `SELECT name, company_id FROM roles WHERE id = $1`

	var roleName string
	var companyID *uuid.UUID

	if err := e.pool.QueryRow(ctx, query, roleID).Scan(&roleName, &companyID); err != nil {
		return "", nil, err
	}

	return roleName, companyID, nil
}

func (e *EmployeeRepository) GetEmployeeWithRoleByEmail(ctx context.Context, email string) (*models.Employee, string, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		SELECT
			e.id, e.company_id, e.email, e.password_hash, e.phone, e.first_name, e.last_name,
			e.employee_code, e.department_id, e.designation_id, e.level_id, e.manager_id,
			e.role_id, e.status, e.employment_type, e.hire_date, e.termination_date,
			e.date_of_birth, e.gender, e.address, e.emergency_contact_name,
			e.emergency_contact_phone, e.profile_image_url, e.last_login_at,
			e.created_at, e.updated_at,
			r.name
		FROM employees e
		JOIN roles r ON r.id = e.role_id
		WHERE e.email = $1
	`

	var employee models.Employee
	var roleName string

	err := e.pool.QueryRow(ctx, query, email).Scan(
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
		&roleName,
	)
	if err != nil {
		return nil, "", err
	}

	return &employee, roleName, nil
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

	if listRequest.DepartmentID != "" {
		where += fmt.Sprintf(" AND department_id = $%d", i)
		args = append(args, listRequest.DepartmentID)
		i++
	}

	if listRequest.ManagerID != "" {
		where += fmt.Sprintf(" AND manager_id = $%d", i)
		args = append(args, listRequest.ManagerID)
		i++
	}

	if listRequest.EmploymentType != "" {
		where += fmt.Sprintf(" AND employment_type = $%d", i)
		args = append(args, listRequest.EmploymentType)
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

func (e *EmployeeRepository) UpdateEmployee(
	ctx context.Context,
	employeeID uuid.UUID,
	employee *models.Employee,
) (*models.Employee, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		UPDATE employees
		SET
			phone = COALESCE(NULLIF($1, ''), phone),
			first_name = COALESCE(NULLIF($2, ''), first_name),
			last_name = COALESCE(NULLIF($3, ''), last_name),
			date_of_birth = CASE WHEN $4::date IS DISTINCT FROM NULL THEN $4::date ELSE date_of_birth END,
			department_id = CASE WHEN $5::uuid IS DISTINCT FROM NULL THEN $5::uuid ELSE department_id END,
			designation_id = CASE WHEN $6::uuid IS DISTINCT FROM NULL THEN $6::uuid ELSE designation_id END,
			level_id = CASE WHEN $7::uuid IS DISTINCT FROM NULL THEN $7::uuid ELSE level_id END,
			manager_id = CASE WHEN $8::uuid IS DISTINCT FROM NULL THEN $8::uuid ELSE manager_id END,
			status = COALESCE(NULLIF($9, ''), status),
			gender = COALESCE(NULLIF($10, ''), gender),
			address = COALESCE(NULLIF($11, ''), address),
			emergency_contact_name = COALESCE(NULLIF($12, ''), emergency_contact_name),
			emergency_contact_phone = COALESCE(NULLIF($13, ''), emergency_contact_phone),
			profile_image_url = COALESCE(NULLIF($14, ''), profile_image_url),
			termination_date = CASE WHEN $15::date IS DISTINCT FROM NULL THEN $15::date ELSE termination_date END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $16
		RETURNING id, company_id, email, password_hash, phone, first_name, last_name,
				  employee_code, department_id, designation_id, level_id, manager_id,
				  role_id, status, employment_type, hire_date, termination_date,
				  date_of_birth, gender, address, emergency_contact_name,
				  emergency_contact_phone, profile_image_url, last_login_at,
				  created_at, updated_at
	`

	var updated models.Employee
	err := e.pool.QueryRow(ctx, query,
		employee.Phone,
		employee.FirstName,
		employee.LastName,
		employee.DateOfBirth,
		employee.DepartmentID,
		employee.DesignationID,
		employee.LevelID,
		employee.ManagerID,
		employee.Status,
		employee.Gender,
		employee.Address,
		employee.EmergencyContactName,
		employee.EmergencyContactPhone,
		employee.ProfileImageURL,
		employee.TerminationDate,
		employeeID,
	).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Email,
		&updated.PasswordHash,
		&updated.Phone,
		&updated.FirstName,
		&updated.LastName,
		&updated.EmployeeCode,
		&updated.DepartmentID,
		&updated.DesignationID,
		&updated.LevelID,
		&updated.ManagerID,
		&updated.RoleID,
		&updated.Status,
		&updated.EmploymentType,
		&updated.HireDate,
		&updated.TerminationDate,
		&updated.DateOfBirth,
		&updated.Gender,
		&updated.Address,
		&updated.EmergencyContactName,
		&updated.EmergencyContactPhone,
		&updated.ProfileImageURL,
		&updated.LastLoginAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
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
