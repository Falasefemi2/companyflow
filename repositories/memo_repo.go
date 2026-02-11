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

type IMemoRepository interface {
	CreateMemo(ctx context.Context, memo *models.Memo, recipientIDs []uuid.UUID) (*models.Memo, error)
	GetMemoByID(ctx context.Context, memoID uuid.UUID) (*models.Memo, error)
	GetMemoList(ctx context.Context, req *dto.MemoListRequest) (*utils.PaginatedResponse[*models.Memo], error)
	MarkMemoRead(ctx context.Context, memoID uuid.UUID, employeeID uuid.UUID) error
	ApproveMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error)
	RejectMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error)
}

type MemoRepository struct {
	pool *pgxpool.Pool
}

func NewMemoRepository(pool *pgxpool.Pool) *MemoRepository {
	return &MemoRepository{pool: pool}
}

func (r *MemoRepository) CreateMemo(ctx context.Context, memo *models.Memo, recipientIDs []uuid.UUID) (*models.Memo, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO memos (
		company_id, employee_id, memo_type, title, content, reference_number, status, current_step, priority
	)
	VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9)
	RETURNING id, company_id, employee_id, memo_type, title, content, reference_number, status, current_step, priority, created_at, updated_at
	`

	err = tx.QueryRow(
		ctx,
		query,
		memo.CompanyID,
		memo.EmployeeID,
		memo.MemoType,
		memo.Title,
		memo.Content,
		memo.ReferenceNumber,
		memo.Status,
		memo.CurrentStep,
		memo.Priority,
	).Scan(
		&memo.ID,
		&memo.CompanyID,
		&memo.EmployeeID,
		&memo.MemoType,
		&memo.Title,
		&memo.Content,
		&memo.ReferenceNumber,
		&memo.Status,
		&memo.CurrentStep,
		&memo.Priority,
		&memo.CreatedAt,
		&memo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	for _, recipientID := range recipientIDs {
		_, err = tx.Exec(
			ctx,
			"INSERT INTO memo_recipients (memo_id, employee_id) VALUES ($1, $2) ON CONFLICT (memo_id, employee_id) DO NOTHING",
			memo.ID,
			recipientID,
		)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return memo, nil
}

func (r *MemoRepository) GetMemoByID(ctx context.Context, memoID uuid.UUID) (*models.Memo, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, company_id, employee_id, memo_type, title, content, reference_number, status, current_step, priority, created_at, updated_at
	FROM memos
	WHERE id = $1
	`

	var memo models.Memo
	var reference sql.NullString
	err := r.pool.QueryRow(ctx, query, memoID).Scan(
		&memo.ID,
		&memo.CompanyID,
		&memo.EmployeeID,
		&memo.MemoType,
		&memo.Title,
		&memo.Content,
		&reference,
		&memo.Status,
		&memo.CurrentStep,
		&memo.Priority,
		&memo.CreatedAt,
		&memo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if reference.Valid {
		memo.ReferenceNumber = reference.String
	}

	return &memo, nil
}

func (r *MemoRepository) GetMemoList(ctx context.Context, req *dto.MemoListRequest) (*utils.PaginatedResponse[*models.Memo], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	where := "WHERE company_id = $1"
	args := []any{req.CompanyID}
	i := 2

	if req.EmployeeID != uuid.Nil {
		where += fmt.Sprintf(" AND employee_id = $%d", i)
		args = append(args, req.EmployeeID)
		i++
	}

	if req.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, req.Status)
		i++
	}

	if req.MemoType != "" {
		where += fmt.Sprintf(" AND memo_type = $%d", i)
		args = append(args, req.MemoType)
		i++
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM memos %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
	SELECT id, company_id, employee_id, memo_type, title, content, reference_number, status, current_step, priority, created_at, updated_at
	FROM memos
	%s
	ORDER BY created_at DESC
	LIMIT $%d OFFSET $%d
	`, where, i, i+1)

	args = append(args, req.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memos := make([]*models.Memo, 0)
	for rows.Next() {
		var memo models.Memo
		var reference sql.NullString
		if err := rows.Scan(
			&memo.ID,
			&memo.CompanyID,
			&memo.EmployeeID,
			&memo.MemoType,
			&memo.Title,
			&memo.Content,
			&reference,
			&memo.Status,
			&memo.CurrentStep,
			&memo.Priority,
			&memo.CreatedAt,
			&memo.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if reference.Valid {
			memo.ReferenceNumber = reference.String
		}
		memos = append(memos, &memo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	return &utils.PaginatedResponse[*models.Memo]{
		Data:       memos,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}, nil
}

func (r *MemoRepository) MarkMemoRead(ctx context.Context, memoID uuid.UUID, employeeID uuid.UUID) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	UPDATE memo_recipients
	SET is_read = true, read_at = CURRENT_TIMESTAMP
	WHERE memo_id = $1 AND employee_id = $2
	`
	result, err := r.pool.Exec(ctx, query, memoID, employeeID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("memo recipient not found")
	}
	return nil
}

func (r *MemoRepository) ApproveMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error) {
	return r.applyMemoDecision(ctx, memoID, approverID, "approved", comments)
}

func (r *MemoRepository) RejectMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error) {
	return r.applyMemoDecision(ctx, memoID, approverID, "rejected", comments)
}

func (r *MemoRepository) applyMemoDecision(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, action string, comments string) (*models.Memo, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var memo models.Memo
	var reference sql.NullString
	err = tx.QueryRow(
		ctx,
		`SELECT id, company_id, employee_id, memo_type, title, content, reference_number, status, current_step, priority, created_at, updated_at
		 FROM memos WHERE id = $1 FOR UPDATE`,
		memoID,
	).Scan(
		&memo.ID, &memo.CompanyID, &memo.EmployeeID, &memo.MemoType, &memo.Title, &memo.Content,
		&reference, &memo.Status, &memo.CurrentStep, &memo.Priority, &memo.CreatedAt, &memo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if reference.Valid {
		memo.ReferenceNumber = reference.String
	}

	if memo.Status != "pending" {
		return nil, fmt.Errorf("memo is not pending. current status: %s", memo.Status)
	}

	workflowSteps, err := r.getWorkflowStepsForMemo(ctx, tx, memo.CompanyID, memo.EmployeeID)
	if err != nil {
		return nil, err
	}

	currentStep := memo.CurrentStep
	finalStep := 1
	if len(workflowSteps) > 0 {
		for _, step := range workflowSteps {
			if step.Step > finalStep {
				finalStep = step.Step
			}
		}
	}

	nextStep := currentStep + 1
	newStatus := memo.Status
	if action == "rejected" {
		newStatus = "rejected"
	} else if currentStep >= finalStep {
		newStatus = "approved"
	} else {
		newStatus = "pending"
	}

	if action == "approved" {
		stepToPersist := nextStep
		if newStatus == "approved" {
			stepToPersist = currentStep
		}
		_, err = tx.Exec(
			ctx,
			`UPDATE memos SET status = $1, current_step = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`,
			newStatus,
			stepToPersist,
			memoID,
		)
	} else {
		_, err = tx.Exec(
			ctx,
			`UPDATE memos SET status = 'rejected', updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
			memoID,
		)
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO approval_history (entity_type, entity_id, step_number, approver_id, action, comments)
		 VALUES ('memo', $1, $2, $3, $4, $5)`,
		memoID,
		currentStep,
		approverID,
		action,
		comments,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetMemoByID(ctx, memoID)
}

func (r *MemoRepository) getWorkflowStepsForMemo(ctx context.Context, tx pgxTx, companyID uuid.UUID, employeeID uuid.UUID) ([]ApprovalWorkflowStep, error) {
	var departmentID *uuid.UUID
	err := tx.QueryRow(ctx, "SELECT department_id FROM employees WHERE id = $1", employeeID).Scan(&departmentID)
	if err != nil {
		return nil, err
	}

	stepsJSON, err := lookupWorkflowSteps(ctx, tx, companyID, "memo", departmentID)
	if err != nil {
		return nil, nil // memo can still be approved without a configured workflow
	}

	steps := make([]ApprovalWorkflowStep, 0)
	if len(stepsJSON) == 0 {
		return steps, nil
	}

	if err := json.Unmarshal(stepsJSON, &steps); err != nil {
		return nil, fmt.Errorf("invalid memo approval workflow steps: %w", err)
	}

	return steps, nil
}
