package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/utils"
)

type ApprovalWorkflow struct {
	ID           uuid.UUID     `db:"id"`
	CompanyID    uuid.UUID     `db:"company_id"`
	WorkflowType string        `db:"workflow_type"`
	DepartmentID *uuid.UUID    `db:"department_id"`
	Steps        utils.JSONRaw `db:"steps"`
	IsActive     bool          `db:"is_active"`
	CreatedAt    time.Time     `db:"created_at"`
}

type ApprovalHistory struct {
	ID         uuid.UUID `db:"id"`
	EntityType string    `db:"entity_type"`
	EntityID   uuid.UUID `db:"entity_id"`
	StepNumber int       `db:"step_number"`
	ApproverID uuid.UUID `db:"approver_id"`
	Action     string    `db:"action"`
	Comments   string    `db:"comments"`
	CreatedAt  time.Time `db:"created_at"`
}
