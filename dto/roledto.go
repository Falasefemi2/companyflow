package dto

import "time"

type CreateRoleRequest struct {
	Name             string   `json:"name" validate:"required,min=2,max=50"`
	Description      string   `json:"description" validate:"omitempty"`
	PermissionsCache []string `json:"permissions_cache" validate:"omitempty"`
}

type UpdateRoleRequest struct {
	Name             *string   `json:"name" validate:"omitempty,min=2,max=50"`
	Description      *string   `json:"description" validate:"omitempty"`
	PermissionsCache *[]string `json:"permissions_cache" validate:"omitempty"`
}

type RoleResponse struct {
	ID               string    `json:"id"`
	CompanyID        *string   `json:"company_id"`
	Name             string    `json:"name"`
	Description      *string   `json:"description"`
	IsSystemRole     bool      `json:"is_system_role"`
	PermissionsCache []string  `json:"permissions_cache"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RoleListRequest struct {
	Page     int    `query:"page" default:"1"`
	PageSize int    `query:"page_size" default:"10"`
	Search   string `query:"search"`
}
