package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaveBalance struct {
	ID                 uuid.UUID `db:"id"`
	EmployeeID         uuid.UUID `db:"employee_id"`
	LeaveTypeID        uuid.UUID `db:"leave_type_id"`
	Year               int       `db:"year"`
	TotalDays          float64   `db:"total_days"`
	UsedDays           float64   `db:"used_days"`
	PendingDays        float64   `db:"pending_days"`
	CarriedForwardDays float64   `db:"carried_forward_days"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

type LeaveRequest struct {
	ID              uuid.UUID  `db:"id"`
	EmployeeID      uuid.UUID  `db:"employee_id"`
	LeaveTypeID     uuid.UUID  `db:"leave_type_id"`
	StartDate       time.Time  `db:"start_date"`
	EndDate         time.Time  `db:"end_date"`
	DaysRequested   float64    `db:"days_requested"`
	Reason          string     `db:"reason"`
	AttachmentURL   *string    `db:"attachment_url"`
	Status          string     `db:"status"` // "pending", "approved", etc.
	CurrentStep     int        `db:"current_step"`
	ApprovedBy      *uuid.UUID `db:"approved_by"`
	ApprovedAt      *time.Time `db:"approved_at"`
	RejectionReason string     `db:"rejection_reason"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
