package models

import (
	"time"

	"github.com/google/uuid"
)

type Department struct {
	ID                 uuid.UUID  `db:"id"`
	CompanyID          uuid.UUID  `db:"company_id"`
	Name               string     `db:"name"`
	Code               string     `db:"code"`
	Description        string     `db:"description"`
	ParentDepartmentID *uuid.UUID `db:"parent_department_id"`
	CostCenter         string     `db:"cost_center"`
	Status             string     `db:"status"` // active, inactive
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}
