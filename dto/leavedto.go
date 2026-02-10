package dto

import "github.com/google/uuid"

type LeaveTypeListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
	Status   string `json:"status"`
}

type CreateLeaveTypeRequest struct {
	Name                  string  `json:"name"`
	Code                  string  `json:"code"`
	Description           string  `json:"description"`
	DaysAllowed           float64 `json:"daysAllowed"`
	IsPaid                bool    `json:"isPaid"`
	RequiresDocumentation bool    `json:"requiresDocumentation"`
	CarryForwardAllowed   bool    `json:"carryForwardAllowed"`
	MaxCarryForwardDays   float64 `json:"maxCarryForwardDays"`
	ColorCode             string  `json:"colorCode"`
	Status                string  `json:"status"`
}

type UpdateLeaveTypeRequest struct {
	Name                  string  `json:"name"`
	Code                  string  `json:"code"`
	Description           string  `json:"description"`
	DaysAllowed           float64 `json:"daysAllowed"`
	IsPaid                *bool   `json:"isPaid"`
	RequiresDocumentation *bool   `json:"requiresDocumentation"`
	CarryForwardAllowed   *bool   `json:"carryForwardAllowed"`
	MaxCarryForwardDays   float64 `json:"maxCarryForwardDays"`
	ColorCode             string  `json:"colorCode"`
	Status                string  `json:"status"`
}

type RequestLeaveRequest struct {
	LeaveTypeID   uuid.UUID `json:"leaveTypeId"`
	StartDate     string    `json:"startDate"` // "2025-02-10" (YYYY-MM-DD format)
	EndDate       string    `json:"endDate"`   // "2025-02-15"
	DaysRequested float64   `json:"daysRequested"`
	Reason        string    `json:"reason"`
	Attachment    string    `json:"attachment"` // Optional: URL
}

type LeaveRequestListRequest struct {
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
	EmployeeID uuid.UUID `json:"employeeId"`
	Status     string    `json:"status"` // pending, approved, rejected, etc.
}

type ApproveLeaveRequestRequest struct {
	RequestID  uuid.UUID `json:"requestId"`
	ApprovedBy uuid.UUID `json:"approvedBy"` // Manager's ID
}

type RejectLeaveRequestRequest struct {
	RequestID       uuid.UUID `json:"requestId"`
	RejectionReason string    `json:"rejectionReason"`
}

type WithdrawLeaveRequestRequest struct {
	RequestID uuid.UUID `json:"requestId"`
}

type LeaveBalanceListRequest struct {
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
	EmployeeID uuid.UUID `json:"employeeId"`
	Year       int       `json:"year"`
}

type CheckBalanceRequest struct {
	LeaveTypeID uuid.UUID `json:"leaveTypeId"`
	Year        int       `json:"year"`
}

type CheckBalanceResponse struct {
	EmployeeID    uuid.UUID `json:"employeeId"`
	LeaveTypeID   uuid.UUID `json:"leaveTypeId"`
	Year          int       `json:"year"`
	AvailableDays float64   `json:"availableDays"`
	TotalDays     float64   `json:"totalDays"`
	UsedDays      float64   `json:"usedDays"`
	PendingDays   float64   `json:"pendingDays"`
}
