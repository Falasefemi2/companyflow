package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgxTx interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func lookupWorkflowSteps(
	ctx context.Context,
	tx pgxTx,
	companyID uuid.UUID,
	workflowType string,
	departmentID *uuid.UUID,
) ([]byte, error) {
	var steps []byte
	var err error

	if departmentID != nil {
		err = tx.QueryRow(
			ctx,
			`SELECT steps
			 FROM approval_workflows
			 WHERE company_id = $1 AND workflow_type = $2 AND department_id = $3 AND is_active = true
			 ORDER BY created_at DESC
			 LIMIT 1`,
			companyID,
			workflowType,
			*departmentID,
		).Scan(&steps)
	}

	if err != nil || len(steps) == 0 {
		err = tx.QueryRow(
			ctx,
			`SELECT steps
			 FROM approval_workflows
			 WHERE company_id = $1 AND workflow_type = $2 AND department_id IS NULL AND is_active = true
			 ORDER BY created_at DESC
			 LIMIT 1`,
			companyID,
			workflowType,
		).Scan(&steps)
	}

	return steps, err
}
