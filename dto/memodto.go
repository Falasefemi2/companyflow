package dto

import "github.com/google/uuid"

type CreateMemoRequest struct {
	CompanyID       uuid.UUID   `json:"companyId"`
	MemoType        string      `json:"memoType"`
	Title           string      `json:"title"`
	Content         string      `json:"content"`
	ReferenceNumber string      `json:"referenceNumber"`
	Priority        string      `json:"priority"`
	RecipientIDs    []uuid.UUID `json:"recipientIds"`
}

type MemoListRequest struct {
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
	CompanyID  uuid.UUID `json:"companyId"`
	EmployeeID uuid.UUID `json:"employeeId"`
	Status     string    `json:"status"`
	MemoType   string    `json:"memoType"`
}

type MemoActionRequest struct {
	Comments string `json:"comments"`
}
