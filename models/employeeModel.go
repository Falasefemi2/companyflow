package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID                    uuid.UUID  `db:"id"`
	CompanyID             uuid.UUID  `db:"company_id"`
	Email                 string     `db:"email"`
	PasswordHash          string     `db:"password_hash"`
	Phone                 string     `db:"phone"`
	FirstName             string     `db:"first_name"`
	LastName              string     `db:"last_name"`
	EmployeeCode          string     `db:"employee_code"`
	DepartmentID          *uuid.UUID `db:"department_id"`
	DesignationID         *uuid.UUID `db:"designation_id"`
	LevelID               *uuid.UUID `db:"level_id"`
	ManagerID             *uuid.UUID `db:"manager_id"`
	RoleID                uuid.UUID  `db:"role_id"`
	Status                string     `db:"status"`
	EmploymentType        string     `db:"employment_type"`
	HireDate              time.Time  `db:"hire_date"`
	TerminationDate       *time.Time `db:"termination_date"`
	DateOfBirth           *time.Time `db:"date_of_birth"`
	Gender                string     `db:"gender"`
	Address               string     `db:"address"`
	EmergencyContactName  string     `db:"emergency_contact_name"`
	EmergencyContactPhone string     `db:"emergency_contact_phone"`
	ProfileImageURL       string     `db:"profile_image_url"`
	LastLoginAt           *time.Time `db:"last_login_at"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
}
