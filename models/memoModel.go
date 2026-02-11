package models

import (
	"time"

	"github.com/google/uuid"
)

type Memo struct {
	ID              uuid.UUID `db:"id"`
	CompanyID       uuid.UUID `db:"company_id"`
	EmployeeID      uuid.UUID `db:"employee_id"`
	MemoType        string    `db:"memo_type"`
	Title           string    `db:"title"`
	Content         string    `db:"content"`
	ReferenceNumber string    `db:"reference_number"`
	Status          string    `db:"status"`
	CurrentStep     int       `db:"current_step"`
	Priority        string    `db:"priority"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type MemoRecipient struct {
	ID         uuid.UUID  `db:"id"`
	MemoID     uuid.UUID  `db:"memo_id"`
	EmployeeID uuid.UUID  `db:"employee_id"`
	IsRead     bool       `db:"is_read"`
	ReadAt     *time.Time `db:"read_at"`
	CreatedAt  time.Time  `db:"created_at"`
}
