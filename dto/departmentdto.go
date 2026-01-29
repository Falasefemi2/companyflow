package dto

import (
	"time"

	"github.com/falasefemi2/companyflowlow/utils"
)

type CreateDepartmentRequest struct {
	Name               string `json:"name" validate:"required,min=2,max=255"`
	Code               string `json:"code" validate:"omitempty,max=50"`
	Description        string `json:"description" validate:"omitempty"`
	ParentDepartmentID string `json:"parent_department_id" validate:"omitempty,uuid"` // For hierarchical departments
	CostCenter         string `json:"cost_center" validate:"omitempty,max=100"`
	Status             string `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateDepartmentRequest struct {
	Name               *string `json:"name" validate:"omitempty,min=2,max=255"`
	Code               *string `json:"code" validate:"omitempty,max=50"`
	Description        *string `json:"description" validate:"omitempty"`
	ParentDepartmentID *string `json:"parent_department_id" validate:"omitempty,uuid"`
	CostCenter         *string `json:"cost_center" validate:"omitempty,max=100"`
	Status             *string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type DepartmentResponse struct {
	ID                 string    `json:"id"`
	CompanyID          string    `json:"company_id"`
	Name               string    `json:"name"`
	Code               string    `json:"code"`
	Description        string    `json:"description"`
	ParentDepartmentID *string   `json:"parent_department_id"`
	CostCenter         string    `json:"cost_center"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DepartmentListRequest struct {
	utils.PaginationParams
	Status string `json:"status" validate:"omitempty,oneof=active inactive"`
	Search string `json:"search" validate:"omitempty"` // Search by name or code
}
