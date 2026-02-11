package dto

import (
	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/utils"
)

type CreateApprovalWorkflowRequest struct {
	CompanyID    uuid.UUID     `json:"companyId"`
	WorkflowType string        `json:"workflowType"`
	DepartmentID *uuid.UUID    `json:"departmentId"`
	Steps        utils.JSONRaw `json:"steps"`
	IsActive     *bool         `json:"isActive"`
}

type ApprovalWorkflowListRequest struct {
	CompanyID    uuid.UUID `json:"companyId"`
	WorkflowType string    `json:"workflowType"`
	DepartmentID uuid.UUID `json:"departmentId"`
	OnlyActive   bool      `json:"onlyActive"`
}

type ApprovalHistoryListRequest struct {
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
}
