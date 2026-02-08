package dto

import (
	"time"
)

type PermissionResponse struct {
	ID         string                 `json:"id"`
	RoleID     string                 `json:"role_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

type CreatePermissionRequest struct {
	Action     string                 `json:"action" binding:"required"`
	Resource   string                 `json:"resource" binding:"required"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

type UpdatePermissionRequest struct {
	Action     *string                 `json:"action,omitempty"`
	Resource   *string                 `json:"resource,omitempty"`
	Conditions *map[string]interface{} `json:"conditions,omitempty"`
}
