package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaveType struct {
	ID                    uuid.UUID `db:"id"`
	CompanyID             uuid.UUID `db:"company_id"`
	Name                  string    `db:"name"`
	Code                  string    `db:"code"`
	Description           string    `db:"description"`
	DaysAllowed           float64   `db:"days_allowed"`
	IsPaid                bool      `db:"is_paid"`
	RequiresDocumentation bool      `db:"requires_documentation"`
	CarryForwardAllowed   bool      `db:"carry_forward_allowed"`
	MaxCarryForwardDays   float64   `db:"max_carry_forward_days"`
	ColorCode             string    `db:"color_code"`
	Status                string    `db:"status"` // "active", "inactive"
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}
