package dto

import (
	"time"

	"github.com/falasefemi2/companyflowlow/utils"
)

type CreateLevelRequest struct {
	Name           string   `json:"name" validate:"required,min=2,max=255"`
	HierarchyLevel int      `json:"hierarchy_level" validate:"required,min=1"`
	MinSalary      *float64 `json:"min_salary" validate:"omitempty,min=0"`
	MaxSalary      *float64 `json:"max_salary" validate:"omitempty,min=0"`
	Description    string   `json:"description" validate:"omitempty"`
}

type UpdateLevelRequest struct {
	Name           *string  `json:"name" validate:"omitempty,min=2,max=255"`
	HierarchyLevel *int     `json:"hierarchy_level" validate:"omitempty,min=1"`
	MinSalary      *float64 `json:"min_salary" validate:"omitempty,min=0"`
	MaxSalary      *float64 `json:"max_salary" validate:"omitempty,min=0"`
	Description    *string  `json:"description" validate:"omitempty"`
}

type LevelResponse struct {
	ID             string    `json:"id"`
	CompanyID      string    `json:"company_id"`
	Name           string    `json:"name"`
	HierarchyLevel int       `json:"hierarchy_level"`
	MinSalary      *float64  `json:"min_salary"`
	MaxSalary      *float64  `json:"max_salary"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LevelListRequest struct {
	utils.PaginationParams
	Search string `json:"search" validate:"omitempty"` // Search by name
}
