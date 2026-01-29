package models

import (
	"time"

	"github.com/google/uuid"
)

type Designation struct {
	ID           uuid.UUID  `db:"id"`
	CompanyID    uuid.UUID  `db:"company_id"`
	Name         string     `db:"name"`
	Description  string     `db:"description"`
	LevelID      *uuid.UUID `db:"level_id"`
	DepartmentID *uuid.UUID `db:"department_id"`
	Status       string     `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}
