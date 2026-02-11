package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/utils"
)

type ILeaveRepository interface {
	CreateLeaveType(ctx context.Context, leaveType *models.LeaveType) (*models.LeaveType, error)

	GetLeaveTypeByID(ctx context.Context, leaveTypeID uuid.UUID) (*models.LeaveType, error)

	GetLeaveTypeList(
		ctx context.Context,
		companyID uuid.UUID,
		listRequest *dto.LeaveTypeListRequest,
	) (*utils.PaginatedResponse[*models.LeaveType], error)

	UpdateLeaveType(ctx context.Context, leaveTypeID uuid.UUID, leaveType *models.LeaveType) (*models.LeaveType, error)

	DeleteLeaveType(ctx context.Context, leaveTypeID uuid.UUID) error

	CreateLeaveRequest(
		ctx context.Context,
		employeeID uuid.UUID,
		leaveTypeID uuid.UUID,
		startDate string, // "2025-02-10"
		endDate string, // "2025-02-15"
		daysRequested float64,
		reason string,
		attachmentURL *string,
	) (*models.LeaveRequest, error)

	GetLeaveRequestByID(ctx context.Context, requestID uuid.UUID) (*models.LeaveRequest, error)

	GetLeaveRequestList(
		ctx context.Context,
		listRequest *dto.LeaveRequestListRequest,
	) (*utils.PaginatedResponse[*models.LeaveRequest], error)

	ApproveLeaveRequest(
		ctx context.Context,
		requestID uuid.UUID,
		approvedByID uuid.UUID,
	) (*models.LeaveRequest, error)

	RejectLeaveRequest(
		ctx context.Context,
		requestID uuid.UUID,
		rejectionReason string,
		rejectedByID ...uuid.UUID,
	) (*models.LeaveRequest, error)

	WithdrawLeaveRequest(ctx context.Context, requestID uuid.UUID, employeeID uuid.UUID) (*models.LeaveRequest, error)

	CheckBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (float64, error)

	GetLeaveBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*models.LeaveBalance, error)

	GetEmployeeBalances(
		ctx context.Context,
		employeeID uuid.UUID,
		year int,
	) ([]*models.LeaveBalance, error)
}

type LeaveRepository struct {
	pool *pgxpool.Pool
}

func NewLeaveRepository(pool *pgxpool.Pool) *LeaveRepository {
	return &LeaveRepository{
		pool: pool,
	}
}

func (l *LeaveRepository) CreateLeaveType(ctx context.Context, leaveType *models.LeaveType) (*models.LeaveType, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	INSERT INTO leave_types (
		company_id, name, code, description, days_allowed, 
		is_paid, requires_documentation, carry_forward_allowed, 
		max_carry_forward_days, color_code, status
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	)
	RETURNING id, company_id, name, code, description, days_allowed, 
		is_paid, requires_documentation, carry_forward_allowed, 
		max_carry_forward_days, color_code, status, created_at, updated_at
	`

	err := l.pool.QueryRow(ctx, query,
		leaveType.CompanyID,
		leaveType.Name,
		leaveType.Code,
		leaveType.Description,
		leaveType.DaysAllowed,
		leaveType.IsPaid,
		leaveType.RequiresDocumentation,
		leaveType.CarryForwardAllowed,
		leaveType.MaxCarryForwardDays,
		leaveType.ColorCode,
		leaveType.Status,
	).Scan(
		&leaveType.ID,
		&leaveType.CompanyID,
		&leaveType.Name,
		&leaveType.Code,
		&leaveType.Description,
		&leaveType.DaysAllowed,
		&leaveType.IsPaid,
		&leaveType.RequiresDocumentation,
		&leaveType.CarryForwardAllowed,
		&leaveType.MaxCarryForwardDays,
		&leaveType.ColorCode,
		&leaveType.Status,
		&leaveType.CreatedAt,
		&leaveType.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return leaveType, nil
}

func (r *LeaveRepository) GetLeaveTypeByID(ctx context.Context, leaveTypeID uuid.UUID) (*models.LeaveType, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, company_id, name, code, description, days_allowed, 
		is_paid, requires_documentation, carry_forward_allowed, 
		max_carry_forward_days, color_code, status, created_at, updated_at
	FROM leave_types
	WHERE id = $1
	`

	var leaveType models.LeaveType

	err := r.pool.QueryRow(ctx, query, leaveTypeID).Scan(
		&leaveType.ID,
		&leaveType.CompanyID,
		&leaveType.Name,
		&leaveType.Code,
		&leaveType.Description,
		&leaveType.DaysAllowed,
		&leaveType.IsPaid,
		&leaveType.RequiresDocumentation,
		&leaveType.CarryForwardAllowed,
		&leaveType.MaxCarryForwardDays,
		&leaveType.ColorCode,
		&leaveType.Status,
		&leaveType.CreatedAt,
		&leaveType.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &leaveType, nil
}

func (r *LeaveRepository) GetLeaveTypeList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.LeaveTypeListRequest,
) (*utils.PaginatedResponse[*models.LeaveType], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Build WHERE clause
	where := "WHERE company_id = $1"
	args := []any{companyID}
	i := 2

	// Search by name if provided
	if listRequest.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", i)
		search := "%" + listRequest.Search + "%"
		args = append(args, search)
		i++
	}

	// Filter by status if provided
	if listRequest.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, listRequest.Status)
		i++
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM leave_types %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (listRequest.Page - 1) * listRequest.PageSize

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, company_id, name, code, description, days_allowed, 
			is_paid, requires_documentation, carry_forward_allowed, 
			max_carry_forward_days, color_code, status, created_at, updated_at
		FROM leave_types
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

	var leaveTypes []*models.LeaveType

	for rows.Next() {
		var lt models.LeaveType
		if err := rows.Scan(
			&lt.ID, &lt.CompanyID, &lt.Name, &lt.Code, &lt.Description,
			&lt.DaysAllowed, &lt.IsPaid, &lt.RequiresDocumentation,
			&lt.CarryForwardAllowed, &lt.MaxCarryForwardDays,
			&lt.ColorCode, &lt.Status, &lt.CreatedAt, &lt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		leaveTypes = append(leaveTypes, &lt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.LeaveType]{
		Data:       leaveTypes,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (r *LeaveRepository) UpdateLeaveType(ctx context.Context, leaveTypeID uuid.UUID, leaveType *models.LeaveType) (*models.LeaveType, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		UPDATE leave_types
		SET
			name = COALESCE(NULLIF($1, ''), name),
			code = COALESCE(NULLIF($2, ''), code),
			description = COALESCE($3, description),
			days_allowed = COALESCE(NULLIF($4, 0), days_allowed),
			is_paid = COALESCE($5, is_paid),
			requires_documentation = COALESCE($6, requires_documentation),
			carry_forward_allowed = COALESCE($7, carry_forward_allowed),
			max_carry_forward_days = COALESCE(NULLIF($8, 0), max_carry_forward_days),
			color_code = COALESCE(NULLIF($9, ''), color_code),
			status = COALESCE(NULLIF($10, ''), status),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
		RETURNING id, company_id, name, code, description, days_allowed, 
			is_paid, requires_documentation, carry_forward_allowed, 
			max_carry_forward_days, color_code, status, created_at, updated_at
	`

	var updated models.LeaveType
	err := r.pool.QueryRow(ctx, query,
		leaveType.Name,
		leaveType.Code,
		leaveType.Description,
		leaveType.DaysAllowed,
		leaveType.IsPaid,
		leaveType.RequiresDocumentation,
		leaveType.CarryForwardAllowed,
		leaveType.MaxCarryForwardDays,
		leaveType.ColorCode,
		leaveType.Status,
		leaveTypeID,
	).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Name,
		&updated.Code,
		&updated.Description,
		&updated.DaysAllowed,
		&updated.IsPaid,
		&updated.RequiresDocumentation,
		&updated.CarryForwardAllowed,
		&updated.MaxCarryForwardDays,
		&updated.ColorCode,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *LeaveRepository) DeleteLeaveType(ctx context.Context, leaveTypeID uuid.UUID) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	result, err := r.pool.Exec(ctx, "DELETE FROM leave_types WHERE id = $1", leaveTypeID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("leave type not found")
	}

	return nil
}

func (l *LeaveRepository) CreateLeaveRequest(
	ctx context.Context,
	employeeID uuid.UUID,
	leaveTypeID uuid.UUID,
	startDate string, // "2025-02-10"
	endDate string, // "2025-02-15"
	daysRequested float64,
	reason string,
	attachmentURL *string,
) (*models.LeaveRequest, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	if end.Before(start) {
		return nil, errors.New("end date cannot be before start date")
	}

	year := start.Year()

	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var leaveTypeExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM leave_types WHERE id = $1 AND status = 'active')",
		leaveTypeID,
	).Scan(&leaveTypeExists)
	if err != nil {
		return nil, err
	}
	if !leaveTypeExists {
		return nil, errors.New("leave type not found or inactive")
	}

	query := `
	SELECT id, total_days, used_days, pending_days
	FROM leave_balances
	WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3
	`
	var balanceID uuid.UUID
	var totalDays, usedDays, pendingDays float64

	err = tx.QueryRow(ctx, query, employeeID, leaveTypeID, year).Scan(
		&balanceID,
		&totalDays,
		&usedDays,
		&pendingDays,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			var daysAllowed float64
			err = tx.QueryRow(ctx,
				"SELECT days_allowed FROM leave_types WHERE id = $1",
				leaveTypeID,
			).Scan(&daysAllowed)
			if err != nil {
				return nil, err
			}

			createBalanceQuery := `
			INSERT INTO leave_balances (employee_id, leave_type_id, year, total_days, used_days, pending_days, carried_forward_days)
			VALUES ($1, $2, $3, $4, 0, 0, 0)
			RETURNING id, total_days, used_days, pending_days
			`
			err = tx.QueryRow(ctx, createBalanceQuery,
				employeeID, leaveTypeID, year, daysAllowed,
			).Scan(&balanceID, &totalDays, &usedDays, &pendingDays)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	availableDays := totalDays - usedDays - pendingDays
	if availableDays < daysRequested {
		return nil, fmt.Errorf("insufficient leave balance. available: %.1f, requested: %.1f", availableDays, daysRequested)
	}

	leaveRequestQuery := `
	INSERT INTO leave_requests (
		employee_id, leave_type_id, start_date, end_date, days_requested, 
		reason, attachment_url, status, current_step, created_at, updated_at
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
	)
	RETURNING id, employee_id, leave_type_id, start_date, end_date, days_requested, 
		reason, attachment_url, status, current_step, approved_by, approved_at, rejection_reason, created_at, updated_at
	`

	var leaveRequest models.LeaveRequest
	var rn sql.NullString

	err = tx.QueryRow(ctx, leaveRequestQuery,
		employeeID,
		leaveTypeID,
		start,
		end,
		daysRequested,
		reason,
		attachmentURL,
		"pending",
		1,
	).Scan(
		&leaveRequest.ID,
		&leaveRequest.EmployeeID,
		&leaveRequest.LeaveTypeID,
		&leaveRequest.StartDate,
		&leaveRequest.EndDate,
		&leaveRequest.DaysRequested,
		&leaveRequest.Reason,
		&leaveRequest.AttachmentURL,
		&leaveRequest.Status,
		&leaveRequest.CurrentStep,
		&leaveRequest.ApprovedBy,
		&leaveRequest.ApprovedAt,
		&rn,
		&leaveRequest.CreatedAt,
		&leaveRequest.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if rn.Valid {
		leaveRequest.RejectionReason = rn.String
	} else {
		leaveRequest.RejectionReason = ""
	}

	updateBalanceQuery := `
	UPDATE leave_balances
	SET pending_days = pending_days + $1, updated_at = CURRENT_TIMESTAMP
	WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateBalanceQuery, daysRequested, balanceID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &leaveRequest, nil
}

func (l *LeaveRepository) GetLeaveRequestByID(ctx context.Context, requestID uuid.UUID) (*models.LeaveRequest, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, employee_id, leave_type_id, start_date, end_date, days_requested,
		reason, attachment_url, status, current_step, approved_by, approved_at,
		rejection_reason, created_at, updated_at
	FROM leave_requests
	WHERE id = $1
	`

	var leaveRequest models.LeaveRequest
	var rn sql.NullString
	err := l.pool.QueryRow(ctx, query, requestID).Scan(
		&leaveRequest.ID,
		&leaveRequest.EmployeeID,
		&leaveRequest.LeaveTypeID,
		&leaveRequest.StartDate,
		&leaveRequest.EndDate,
		&leaveRequest.DaysRequested,
		&leaveRequest.Reason,
		&leaveRequest.AttachmentURL,
		&leaveRequest.Status,
		&leaveRequest.CurrentStep,
		&leaveRequest.ApprovedBy,
		&leaveRequest.ApprovedAt,
		&rn,
		&leaveRequest.CreatedAt,
		&leaveRequest.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if rn.Valid {
		leaveRequest.RejectionReason = rn.String
	} else {
		leaveRequest.RejectionReason = ""
	}

	return &leaveRequest, nil
}

func (l *LeaveRepository) GetLeaveRequestList(
	ctx context.Context,
	listRequest *dto.LeaveRequestListRequest,
) (*utils.PaginatedResponse[*models.LeaveRequest], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	where := "WHERE 1=1"
	args := []any{}
	i := 1

	// Filter by employee if provided
	if listRequest.EmployeeID != uuid.Nil {
		where += fmt.Sprintf(" AND employee_id = $%d", i)
		args = append(args, listRequest.EmployeeID)
		i++
	}

	// Filter by status if provided
	if listRequest.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, listRequest.Status)
		i++
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM leave_requests %s", where)
	if err := l.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (listRequest.Page - 1) * listRequest.PageSize

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, employee_id, leave_type_id, start_date, end_date, days_requested,
			reason, attachment_url, status, current_step, approved_by, approved_at,
			rejection_reason, created_at, updated_at
		FROM leave_requests
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

	var leaveRequests []*models.LeaveRequest

	for rows.Next() {
		var lr models.LeaveRequest
		var rn sql.NullString
		if err := rows.Scan(
			&lr.ID,
			&lr.EmployeeID,
			&lr.LeaveTypeID,
			&lr.StartDate,
			&lr.EndDate,
			&lr.DaysRequested,
			&lr.Reason,
			&lr.AttachmentURL,
			&lr.Status,
			&lr.CurrentStep,
			&lr.ApprovedBy,
			&lr.ApprovedAt,
			&rn,
			&lr.CreatedAt,
			&lr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if rn.Valid {
			lr.RejectionReason = rn.String
		} else {
			lr.RejectionReason = ""
		}
		leaveRequests = append(leaveRequests, &lr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.LeaveRequest]{
		Data:       leaveRequests,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (l *LeaveRepository) ApproveLeaveRequest(
	ctx context.Context,
	requestID uuid.UUID,
	approvedByID uuid.UUID,
) (*models.LeaveRequest, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	getRequestQuery := `
	SELECT
		lr.employee_id,
		lr.leave_type_id,
		lr.days_requested,
		lr.start_date,
		lr.current_step,
		e.company_id,
		e.department_id
	FROM leave_requests lr
	JOIN employees e ON e.id = lr.employee_id
	WHERE lr.id = $1 AND lr.status = 'pending'
	FOR UPDATE
	`

	var employeeID, leaveTypeID, companyID uuid.UUID
	var departmentID *uuid.UUID
	var daysRequested float64
	var startDate time.Time
	var currentStep int

	err = tx.QueryRow(ctx, getRequestQuery, requestID).Scan(
		&employeeID,
		&leaveTypeID,
		&daysRequested,
		&startDate,
		&currentStep,
		&companyID,
		&departmentID,
	)
	if err != nil {
		return nil, err
	}
	if currentStep < 1 {
		currentStep = 1
	}

	year := startDate.Year()

	finalStep := currentStep
	stepsJSON, _ := lookupWorkflowSteps(ctx, tx, companyID, "leave", departmentID)
	if len(stepsJSON) > 0 {
		var workflowSteps []ApprovalWorkflowStep
		if err := json.Unmarshal(stepsJSON, &workflowSteps); err != nil {
			return nil, fmt.Errorf("invalid leave workflow steps: %w", err)
		}
		stepCfg, err := getWorkflowStep(workflowSteps, currentStep)
		if err != nil {
			return nil, err
		}
		if err := validateStepApprover(ctx, tx, stepCfg, approvedByID); err != nil {
			return nil, err
		}
		finalStep = getFinalWorkflowStep(workflowSteps)
	}

	var leaveRequest models.LeaveRequest
	var rn sql.NullString
	if currentStep >= finalStep {
		updateRequestQuery := `
		UPDATE leave_requests
		SET status = 'approved', approved_by = $1, approved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, employee_id, leave_type_id, start_date, end_date, days_requested,
			reason, attachment_url, status, current_step, approved_by, approved_at,
			rejection_reason, created_at, updated_at
		`

		err = tx.QueryRow(ctx, updateRequestQuery, approvedByID, requestID).Scan(
			&leaveRequest.ID,
			&leaveRequest.EmployeeID,
			&leaveRequest.LeaveTypeID,
			&leaveRequest.StartDate,
			&leaveRequest.EndDate,
			&leaveRequest.DaysRequested,
			&leaveRequest.Reason,
			&leaveRequest.AttachmentURL,
			&leaveRequest.Status,
			&leaveRequest.CurrentStep,
			&leaveRequest.ApprovedBy,
			&leaveRequest.ApprovedAt,
			&rn,
			&leaveRequest.CreatedAt,
			&leaveRequest.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		updateBalanceQuery := `
		UPDATE leave_balances
		SET 
			pending_days = pending_days - $1,
			used_days = used_days + $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE employee_id = $3 AND leave_type_id = $4 AND year = $5
		`

		if _, err = tx.Exec(ctx, updateBalanceQuery, daysRequested, daysRequested, employeeID, leaveTypeID, year); err != nil {
			return nil, err
		}
	} else {
		updateRequestQuery := `
		UPDATE leave_requests
		SET current_step = current_step + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, employee_id, leave_type_id, start_date, end_date, days_requested,
			reason, attachment_url, status, current_step, approved_by, approved_at,
			rejection_reason, created_at, updated_at
		`

		err = tx.QueryRow(ctx, updateRequestQuery, requestID).Scan(
			&leaveRequest.ID,
			&leaveRequest.EmployeeID,
			&leaveRequest.LeaveTypeID,
			&leaveRequest.StartDate,
			&leaveRequest.EndDate,
			&leaveRequest.DaysRequested,
			&leaveRequest.Reason,
			&leaveRequest.AttachmentURL,
			&leaveRequest.Status,
			&leaveRequest.CurrentStep,
			&leaveRequest.ApprovedBy,
			&leaveRequest.ApprovedAt,
			&rn,
			&leaveRequest.CreatedAt,
			&leaveRequest.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	if rn.Valid {
		leaveRequest.RejectionReason = rn.String
	} else {
		leaveRequest.RejectionReason = ""
	}

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO approval_history (entity_type, entity_id, step_number, approver_id, action, comments)
		 VALUES ('leave_request', $1, $2, $3, 'approved', $4)`,
		requestID,
		currentStep,
		approvedByID,
		"",
	); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &leaveRequest, nil
}

func (l *LeaveRepository) RejectLeaveRequest(
	ctx context.Context,
	requestID uuid.UUID,
	rejectionReason string,
	rejectedByID ...uuid.UUID,
) (*models.LeaveRequest, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	getRequestQuery := `
	SELECT
		lr.employee_id,
		lr.leave_type_id,
		lr.days_requested,
		lr.start_date,
		lr.current_step,
		e.company_id,
		e.department_id
	FROM leave_requests lr
	JOIN employees e ON e.id = lr.employee_id
	WHERE lr.id = $1 AND lr.status = 'pending'
	FOR UPDATE
	`

	var employeeID, leaveTypeID, companyID uuid.UUID
	var departmentID *uuid.UUID
	var daysRequested float64
	var startDate time.Time
	var currentStep int

	err = tx.QueryRow(ctx, getRequestQuery, requestID).Scan(
		&employeeID,
		&leaveTypeID,
		&daysRequested,
		&startDate,
		&currentStep,
		&companyID,
		&departmentID,
	)
	if err != nil {
		return nil, err
	}
	if currentStep < 1 {
		currentStep = 1
	}

	var actorID uuid.UUID
	if len(rejectedByID) > 0 {
		actorID = rejectedByID[0]
	}

	stepsJSON, _ := lookupWorkflowSteps(ctx, tx, companyID, "leave", departmentID)
	if len(stepsJSON) > 0 && actorID != uuid.Nil {
		var workflowSteps []ApprovalWorkflowStep
		if err := json.Unmarshal(stepsJSON, &workflowSteps); err != nil {
			return nil, fmt.Errorf("invalid leave workflow steps: %w", err)
		}
		stepCfg, err := getWorkflowStep(workflowSteps, currentStep)
		if err != nil {
			return nil, err
		}
		if err := validateStepApprover(ctx, tx, stepCfg, actorID); err != nil {
			return nil, err
		}
	}

	year := startDate.Year()

	updateRequestQuery := `
	UPDATE leave_requests
	SET status = 'rejected', rejection_reason = $1, updated_at = CURRENT_TIMESTAMP
	WHERE id = $2
	RETURNING id, employee_id, leave_type_id, start_date, end_date, days_requested,
		reason, attachment_url, status, current_step, approved_by, approved_at,
		rejection_reason, created_at, updated_at
	`

	var leaveRequest models.LeaveRequest
	var rn sql.NullString
	err = tx.QueryRow(ctx, updateRequestQuery, rejectionReason, requestID).Scan(
		&leaveRequest.ID,
		&leaveRequest.EmployeeID,
		&leaveRequest.LeaveTypeID,
		&leaveRequest.StartDate,
		&leaveRequest.EndDate,
		&leaveRequest.DaysRequested,
		&leaveRequest.Reason,
		&leaveRequest.AttachmentURL,
		&leaveRequest.Status,
		&leaveRequest.CurrentStep,
		&leaveRequest.ApprovedBy,
		&leaveRequest.ApprovedAt,
		&rn,
		&leaveRequest.CreatedAt,
		&leaveRequest.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if rn.Valid {
		leaveRequest.RejectionReason = rn.String
	} else {
		leaveRequest.RejectionReason = ""
	}

	if actorID != uuid.Nil {
		_, err = tx.Exec(
			ctx,
			`INSERT INTO approval_history (entity_type, entity_id, step_number, approver_id, action, comments)
			 VALUES ('leave_request', $1, $2, $3, 'rejected', $4)`,
			requestID,
			currentStep,
			actorID,
			rejectionReason,
		)
		if err != nil {
			return nil, err
		}
	}

	updateBalanceQuery := `
	UPDATE leave_balances
	SET 
		pending_days = pending_days - $1,
		updated_at = CURRENT_TIMESTAMP
	WHERE employee_id = $2 AND leave_type_id = $3 AND year = $4
	`

	_, err = tx.Exec(ctx, updateBalanceQuery, daysRequested, employeeID, leaveTypeID, year)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &leaveRequest, nil
}

func (l *LeaveRepository) WithdrawLeaveRequest(ctx context.Context, requestID uuid.UUID, employeeID uuid.UUID) (*models.LeaveRequest, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	getRequestQuery := `
	SELECT leave_type_id, days_requested, start_date
	FROM leave_requests
	WHERE id = $1 AND employee_id = $2 AND status = 'pending'
	`

	var leaveTypeID uuid.UUID
	var daysRequested float64
	var startDate time.Time

	err = tx.QueryRow(ctx, getRequestQuery, requestID, employeeID).Scan(
		&leaveTypeID,
		&daysRequested,
		&startDate,
	)
	if err != nil {
		return nil, errors.New("request not found or cannot be withdrawn")
	}

	year := startDate.Year()

	updateRequestQuery := `
	UPDATE leave_requests
	SET status = 'withdrawn', updated_at = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING id, employee_id, leave_type_id, start_date, end_date, days_requested,
		reason, attachment_url, status, current_step, approved_by, approved_at,
		rejection_reason, created_at, updated_at
	`

	var leaveRequest models.LeaveRequest
	var rn sql.NullString
	err = tx.QueryRow(ctx, updateRequestQuery, requestID).Scan(
		&leaveRequest.ID,
		&leaveRequest.EmployeeID,
		&leaveRequest.LeaveTypeID,
		&leaveRequest.StartDate,
		&leaveRequest.EndDate,
		&leaveRequest.DaysRequested,
		&leaveRequest.Reason,
		&leaveRequest.AttachmentURL,
		&leaveRequest.Status,
		&leaveRequest.CurrentStep,
		&leaveRequest.ApprovedBy,
		&leaveRequest.ApprovedAt,
		&rn,
		&leaveRequest.CreatedAt,
		&leaveRequest.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if rn.Valid {
		leaveRequest.RejectionReason = rn.String
	} else {
		leaveRequest.RejectionReason = ""
	}

	updateBalanceQuery := `
	UPDATE leave_balances
	SET 
		pending_days = pending_days - $1,
		updated_at = CURRENT_TIMESTAMP
	WHERE employee_id = $2 AND leave_type_id = $3 AND year = $4
	`

	_, err = tx.Exec(ctx, updateBalanceQuery, daysRequested, employeeID, leaveTypeID, year)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &leaveRequest, nil
}

func (l *LeaveRepository) CheckBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (float64, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT total_days, used_days, pending_days
	FROM leave_balances
	WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3
	`

	var totalDays, usedDays, pendingDays float64

	err := l.pool.QueryRow(ctx, query, employeeID, leaveTypeID, year).Scan(&totalDays, &usedDays, &pendingDays)
	if err != nil {
		return 0, nil // No balance found, return 0 available
	}

	available := totalDays - usedDays - pendingDays
	return available, nil
}

func (l *LeaveRepository) GetEmployeeBalances(
	ctx context.Context,
	employeeID uuid.UUID,
	year int,
) ([]*models.LeaveBalance, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, employee_id, leave_type_id, year, total_days, used_days, pending_days,
		carried_forward_days, created_at, updated_at
	FROM leave_balances
	WHERE employee_id = $1 AND year = $2
	ORDER BY created_at DESC
	`

	rows, err := l.pool.Query(ctx, query, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*models.LeaveBalance

	for rows.Next() {
		var lb models.LeaveBalance
		if err := rows.Scan(
			&lb.ID,
			&lb.EmployeeID,
			&lb.LeaveTypeID,
			&lb.Year,
			&lb.TotalDays,
			&lb.UsedDays,
			&lb.PendingDays,
			&lb.CarriedForwardDays,
			&lb.CreatedAt,
			&lb.UpdatedAt,
		); err != nil {
			return nil, err
		}
		balances = append(balances, &lb)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return balances, nil
}

func (l *LeaveRepository) GetLeaveBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*models.LeaveBalance, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, employee_id, leave_type_id, year, total_days, used_days, pending_days,
		carried_forward_days, created_at, updated_at
	FROM leave_balances
	WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3
	`

	var lb models.LeaveBalance
	err := l.pool.QueryRow(ctx, query, employeeID, leaveTypeID, year).Scan(
		&lb.ID,
		&lb.EmployeeID,
		&lb.LeaveTypeID,
		&lb.Year,
		&lb.TotalDays,
		&lb.UsedDays,
		&lb.PendingDays,
		&lb.CarriedForwardDays,
		&lb.CreatedAt,
		&lb.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &lb, nil
}

func getWorkflowStep(steps []ApprovalWorkflowStep, stepNumber int) (*ApprovalWorkflowStep, error) {
	for i := range steps {
		if steps[i].Step == stepNumber {
			return &steps[i], nil
		}
	}
	return nil, fmt.Errorf("no workflow step configured for step %d", stepNumber)
}

func getFinalWorkflowStep(steps []ApprovalWorkflowStep) int {
	finalStep := 1
	for _, step := range steps {
		if step.Step > finalStep {
			finalStep = step.Step
		}
	}
	return finalStep
}

func validateStepApprover(ctx context.Context, tx pgxTx, stepCfg *ApprovalWorkflowStep, approverID uuid.UUID) error {
	if stepCfg == nil {
		return nil
	}

	if stepCfg.ApproverID != nil && *stepCfg.ApproverID != approverID {
		return errors.New("you are not assigned to this approval step")
	}

	if stepCfg.RoleID != nil {
		var approverRoleID uuid.UUID
		if err := tx.QueryRow(ctx, "SELECT role_id FROM employees WHERE id = $1", approverID).Scan(&approverRoleID); err != nil {
			return errors.New("unable to verify approver role")
		}
		if approverRoleID != *stepCfg.RoleID {
			return errors.New("approver role does not match workflow step requirement")
		}
	}

	return nil
}
