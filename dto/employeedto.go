package dto

import (
	"time"

	"github.com/falasefemi2/companyflowlow/utils"
)

type CreateEmployeeRequest struct {
	Email                 string `json:"email" validate:"required,email"`
	Password              string `json:"password" validate:"required,min=8"`
	Phone                 string `json:"phone" validate:"required"`
	FirstName             string `json:"first_name" validate:"required"`
	LastName              string `json:"last_name" validate:"required"`
	DateOfBirth           string `json:"date_of_birth" validate:"required"` // Format: YYYY-MM-DD
	EmployeeCode          string `json:"employee_code" validate:"required"` // Internal ID like EMP001
	DepartmentID          string `json:"department_id" validate:"required,uuid"`
	DesignationID         string `json:"designation_id" validate:"required,uuid"`
	LevelID               string `json:"level_id" validate:"required,uuid"`
	RoleID                string `json:"role_id" validate:"required,uuid"`
	ManagerID             string `json:"manager_id" validate:"omitempty,uuid"` // Optional
	Status                string `json:"status" validate:"required,oneof=active inactive on_leave terminated probation"`
	EmploymentType        string `json:"employment_type" validate:"required,oneof=full_time part_time contract intern"`
	HireDate              string `json:"hire_date" validate:"required"` // Format: YYYY-MM-DD
	Gender                string `json:"gender" validate:"omitempty"`
	Address               string `json:"address" validate:"omitempty"`
	EmergencyContactName  string `json:"emergency_contact_name" validate:"omitempty"`
	EmergencyContactPhone string `json:"emergency_contact_phone" validate:"omitempty"`
	ProfileImageUrl       string `json:"profile_image_url" validate:"omitempty"`
}

type UpdateEmployeeRequest struct {
	Phone                 *string `json:"phone" validate:"omitempty"`
	FirstName             *string `json:"first_name" validate:"omitempty"`
	LastName              *string `json:"last_name" validate:"omitempty"`
	DateOfBirth           *string `json:"date_of_birth" validate:"omitempty"` // Format: YYYY-MM-DD
	DepartmentID          *string `json:"department_id" validate:"omitempty,uuid"`
	DesignationID         *string `json:"designation_id" validate:"omitempty,uuid"`
	LevelID               *string `json:"level_id" validate:"omitempty,uuid"`
	ManagerID             *string `json:"manager_id" validate:"omitempty,uuid"`
	Status                *string `json:"status" validate:"omitempty,oneof=active inactive on_leave terminated probation"`
	Gender                *string `json:"gender" validate:"omitempty"`
	Address               *string `json:"address" validate:"omitempty"`
	EmergencyContactName  *string `json:"emergency_contact_name" validate:"omitempty"`
	EmergencyContactPhone *string `json:"emergency_contact_phone" validate:"omitempty"`
	ProfileImageUrl       *string `json:"profile_image_url" validate:"omitempty"`
	TerminationDate       *string `json:"termination_date" validate:"omitempty"` // Format: YYYY-MM-DD
}

type EmployeeResponse struct {
	ID                    string     `json:"id"`
	CompanyID             string     `json:"company_id"`
	Email                 string     `json:"email"`
	Phone                 string     `json:"phone"`
	FirstName             string     `json:"first_name"`
	LastName              string     `json:"last_name"`
	EmployeeCode          string     `json:"employee_code"`
	DepartmentID          string     `json:"department_id"`
	DesignationID         string     `json:"designation_id"`
	LevelID               string     `json:"level_id"`
	ManagerID             string     `json:"manager_id"`
	RoleID                string     `json:"role_id"`
	Status                string     `json:"status"`
	EmploymentType        string     `json:"employment_type"`
	DateOfBirth           *time.Time `json:"date_of_birth"`
	HireDate              time.Time  `json:"hire_date"`
	TerminationDate       *time.Time `json:"termination_date"`
	Gender                string     `json:"gender"`
	Address               string     `json:"address"`
	EmergencyContactName  string     `json:"emergency_contact_name"`
	EmergencyContactPhone string     `json:"emergency_contact_phone"`
	ProfileImageUrl       string     `json:"profile_image_url"`
	LastLoginAt           *time.Time `json:"last_login_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type EmployeeListRequest struct {
	utils.PaginationParams
	Status         string `json:"status" validate:"omitempty"`          // Filter by status
	DepartmentID   string `json:"department_id" validate:"omitempty"`   // Filter by department
	ManagerID      string `json:"manager_id" validate:"omitempty"`      // Filter by manager
	EmploymentType string `json:"employment_type" validate:"omitempty"` // Filter by employment type
	Search         string `json:"search" validate:"omitempty"`          // Search in name/email
}
