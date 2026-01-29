package dto

import (
	"time"

	"github.com/falasefemi2/companyflowlow/utils"
)

type CreateDesignationRequest struct {
	Name         string `json:"name" validate:"required,min=2,max=255"`
	Description  string `json:"description" validate:"omitempty"`
	LevelID      string `json:"level_id" validate:"omitempty,uuid"`
	DepartmentID string `json:"department_id" validate:"omitempty,uuid"`
	Status       string `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateDesignationRequest struct {
	Name         *string `json:"name" validate:"omitempty,min=2,max=255"`
	Description  *string `json:"description" validate:"omitempty"`
	LevelID      *string `json:"level_id" validate:"omitempty,uuid"`
	DepartmentID *string `json:"department_id" validate:"omitempty,uuid"`
	Status       *string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type DesignationResponse struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"company_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	LevelID      *string   `json:"level_id"`
	DepartmentID *string   `json:"department_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DesignationListRequest struct {
	utils.PaginationParams
	Status       string `json:"status" validate:"omitempty,oneof=active inactive"`
	DepartmentID string `json:"department_id" validate:"omitempty,uuid"`
	LevelID      string `json:"level_id" validate:"omitempty,uuid"`
	Search       string `json:"search" validate:"omitempty"` // Search by name
}
