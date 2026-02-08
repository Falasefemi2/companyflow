package models

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID         uuid.UUID              `db:"id"`
	RoleID     uuid.UUID              `db:"role_id"`
	Action     string                 `db:"action"`
	Resource   string                 `db:"resource"`
	Conditions map[string]interface{} `db:"conditions"`
	CreatedAt  time.Time              `db:"created_at"`
}
