package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
)

type ApprovalWorkflowStep struct {
	Step       int        `json:"step"`
	RoleID     *uuid.UUID `json:"role_id"`
	ApproverID *uuid.UUID `json:"approver_id"`
}

type IApprovalRepository interface {
	CreateWorkflow(ctx context.Context, workflow *models.ApprovalWorkflow) (*models.ApprovalWorkflow, error)
	GetWorkflowList(ctx context.Context, req *dto.ApprovalWorkflowListRequest) ([]*models.ApprovalWorkflow, error)
	GetActiveWorkflow(ctx context.Context, companyID uuid.UUID, workflowType string, departmentID *uuid.UUID) (*models.ApprovalWorkflow, []ApprovalWorkflowStep, error)
	CreateApprovalHistory(ctx context.Context, history *models.ApprovalHistory) (*models.ApprovalHistory, error)
	GetApprovalHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.ApprovalHistory, error)
}

type ApprovalRepository struct {
	pool *pgxpool.Pool
}

func NewApprovalRepository(pool *pgxpool.Pool) *ApprovalRepository {
	return &ApprovalRepository{pool: pool}
}

func (r *ApprovalRepository) CreateWorkflow(ctx context.Context, workflow *models.ApprovalWorkflow) (*models.ApprovalWorkflow, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if !json.Valid(workflow.Steps) {
		return nil, errors.New("workflow steps must be valid JSON")
	}

	if workflow.WorkflowType == "" {
		return nil, errors.New("workflow type is required")
	}

	query := `
	INSERT INTO approval_workflows (company_id, workflow_type, department_id, steps, is_active)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, company_id, workflow_type, department_id, steps, is_active, created_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		workflow.CompanyID,
		workflow.WorkflowType,
		workflow.DepartmentID,
		workflow.Steps,
		workflow.IsActive,
	).Scan(
		&workflow.ID,
		&workflow.CompanyID,
		&workflow.WorkflowType,
		&workflow.DepartmentID,
		&workflow.Steps,
		&workflow.IsActive,
		&workflow.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func (r *ApprovalRepository) GetWorkflowList(ctx context.Context, req *dto.ApprovalWorkflowListRequest) ([]*models.ApprovalWorkflow, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	where := "WHERE company_id = $1"
	args := []any{req.CompanyID}
	i := 2

	if req.WorkflowType != "" {
		where += fmt.Sprintf(" AND workflow_type = $%d", i)
		args = append(args, req.WorkflowType)
		i++
	}

	if req.DepartmentID != uuid.Nil {
		where += fmt.Sprintf(" AND department_id = $%d", i)
		args = append(args, req.DepartmentID)
		i++
	}

	if req.OnlyActive {
		where += " AND is_active = true"
	}

	query := fmt.Sprintf(`
	SELECT id, company_id, workflow_type, department_id, steps, is_active, created_at
	FROM approval_workflows
	%s
	ORDER BY created_at DESC
	`, where)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workflows := make([]*models.ApprovalWorkflow, 0)
	for rows.Next() {
		var wf models.ApprovalWorkflow
		if err := rows.Scan(
			&wf.ID,
			&wf.CompanyID,
			&wf.WorkflowType,
			&wf.DepartmentID,
			&wf.Steps,
			&wf.IsActive,
			&wf.CreatedAt,
		); err != nil {
			return nil, err
		}
		workflows = append(workflows, &wf)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workflows, nil
}

func (r *ApprovalRepository) GetActiveWorkflow(
	ctx context.Context,
	companyID uuid.UUID,
	workflowType string,
	departmentID *uuid.UUID,
) (*models.ApprovalWorkflow, []ApprovalWorkflowStep, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	var wf models.ApprovalWorkflow
	var err error

	if departmentID != nil {
		query := `
		SELECT id, company_id, workflow_type, department_id, steps, is_active, created_at
		FROM approval_workflows
		WHERE company_id = $1 AND workflow_type = $2 AND department_id = $3 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1
		`
		err = r.pool.QueryRow(ctx, query, companyID, workflowType, *departmentID).Scan(
			&wf.ID, &wf.CompanyID, &wf.WorkflowType, &wf.DepartmentID, &wf.Steps, &wf.IsActive, &wf.CreatedAt,
		)
	}

	if err != nil || wf.ID == uuid.Nil {
		query := `
		SELECT id, company_id, workflow_type, department_id, steps, is_active, created_at
		FROM approval_workflows
		WHERE company_id = $1 AND workflow_type = $2 AND department_id IS NULL AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1
		`
		err = r.pool.QueryRow(ctx, query, companyID, workflowType).Scan(
			&wf.ID, &wf.CompanyID, &wf.WorkflowType, &wf.DepartmentID, &wf.Steps, &wf.IsActive, &wf.CreatedAt,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	steps := make([]ApprovalWorkflowStep, 0)
	if err := json.Unmarshal(wf.Steps, &steps); err != nil {
		return nil, nil, fmt.Errorf("invalid workflow steps JSON: %w", err)
	}

	return &wf, steps, nil
}

func (r *ApprovalRepository) CreateApprovalHistory(ctx context.Context, history *models.ApprovalHistory) (*models.ApprovalHistory, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	INSERT INTO approval_history (entity_type, entity_id, step_number, approver_id, action, comments)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, entity_type, entity_id, step_number, approver_id, action, comments, created_at
	`

	if history.Comments == "" {
		history.Comments = ""
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		history.EntityType,
		history.EntityID,
		history.StepNumber,
		history.ApproverID,
		history.Action,
		history.Comments,
	).Scan(
		&history.ID,
		&history.EntityType,
		&history.EntityID,
		&history.StepNumber,
		&history.ApproverID,
		&history.Action,
		&history.Comments,
		&history.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func (r *ApprovalRepository) GetApprovalHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.ApprovalHistory, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
	SELECT id, entity_type, entity_id, step_number, approver_id, action, comments, created_at
	FROM approval_history
	WHERE entity_type = $1 AND entity_id = $2
	ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]*models.ApprovalHistory, 0)
	for rows.Next() {
		var h models.ApprovalHistory
		if err := rows.Scan(&h.ID, &h.EntityType, &h.EntityID, &h.StepNumber, &h.ApproverID, &h.Action, &h.Comments, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}
