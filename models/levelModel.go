package models

import (
	"time"

	"github.com/google/uuid"
)

type Level struct {
	ID             uuid.UUID `db:"id"`
	CompanyID      uuid.UUID `db:"company_id"`
	Name           string    `db:"name"`
	HierarchyLevel int       `db:"hierarchy_level"` // 1, 2, 3, etc. for ordering
	MinSalary      *float64  `db:"min_salary"`
	MaxSalary      *float64  `db:"max_salary"`
	Description    string    `db:"description"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
