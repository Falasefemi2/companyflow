package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID               uuid.UUID  `db:"id"`
	CompanyID        *uuid.UUID `db:"company_id"`
	Name             string     `db:"name"`
	Description      *string    `db:"description"`
	IsSystemRole     bool       `db:"is_system_role"`
	PermissionsCache []string   `db:"permissions_cache"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}
