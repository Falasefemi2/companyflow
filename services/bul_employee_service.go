package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
)

type BulkEmployeeService struct {
	employeeRepo *repositories.EmployeeRepository
	validator    *BulkEmployeeValidator
}

func NewBulkEmployeeService(
	employeeRepo *repositories.EmployeeRepository,
	validator *BulkEmployeeValidator,
) *BulkEmployeeService {
	return &BulkEmployeeService{
		employeeRepo: employeeRepo,
		validator:    validator,
	}
}

// BulkCreateEmployeesRequest combines validation request and company context
type BulkCreateEmployeesRequest struct {
	CompanyID uuid.UUID
	Records   []dto.BulkEmployeeRecord
}

// BulkCreateEmployeesResponse contains results and temp passwords
type BulkCreateEmployeesResponse struct {
	SuccessCount     int
	FailureCount     int
	CreatedEmployees []*BulkCreatedEmployee
	ValidationErrors []dto.BulkEmployeeError
}

// BulkCreatedEmployee contains employee data and their temp password
type BulkCreatedEmployee struct {
	EmployeeID   uuid.UUID
	Email        string
	FirstName    string
	LastName     string
	EmployeeCode string
	TempPassword string // Must be shared with user securely
	CreatedAt    time.Time
}

// CreateBulkEmployees orchestrates the entire bulk creation process
func (s *BulkEmployeeService) CreateBulkEmployees(
	ctx context.Context,
	request *BulkCreateEmployeesRequest,
) (*BulkCreateEmployeesResponse, error) {
	// Step 1: Validate all records
	validationErrors, err := s.validator.ValidateBulkRecords(ctx, request.CompanyID, request.Records)
	if err != nil {
		return nil, fmt.Errorf("validation process failed: %w", err)
	}

	// Step 2: If any validation errors, return them without inserting anything
	if len(validationErrors) > 0 {
		return &BulkCreateEmployeesResponse{
			SuccessCount:     0,
			FailureCount:     len(validationErrors),
			ValidationErrors: validationErrors,
		}, nil
	}

	// Step 3: Transform records to Employee models with hashed passwords
	employees, tempPasswords, err := s.transformRecordsToEmployees(ctx, request.CompanyID, request.Records)
	if err != nil {
		return nil, fmt.Errorf("failed to transform records: %w", err)
	}

	// Step 4: Bulk insert into database (all-or-nothing)
	createdEmployees, err := s.employeeRepo.BulkCreateEmployees(ctx, employees)
	if err != nil {
		return nil, fmt.Errorf("failed to create employees in database: %w", err)
	}

	// Step 5: Combine created employees with their temp passwords
	responseEmployees := make([]*BulkCreatedEmployee, len(createdEmployees))
	for i, emp := range createdEmployees {
		responseEmployees[i] = &BulkCreatedEmployee{
			EmployeeID:   emp.ID,
			Email:        emp.Email,
			FirstName:    emp.FirstName,
			LastName:     emp.LastName,
			EmployeeCode: emp.EmployeeCode,
			TempPassword: tempPasswords[i],
			CreatedAt:    emp.CreatedAt,
		}
	}

	return &BulkCreateEmployeesResponse{
		SuccessCount:     len(createdEmployees),
		FailureCount:     0,
		CreatedEmployees: responseEmployees,
		ValidationErrors: nil,
	}, nil
}

// transformRecordsToEmployees converts DTO records to Employee models
// Also generates and hashes passwords
func (s *BulkEmployeeService) transformRecordsToEmployees(
	ctx context.Context,
	companyID uuid.UUID,
	records []dto.BulkEmployeeRecord,
) ([]*models.Employee, []string, error) {
	employees := make([]*models.Employee, len(records))
	tempPasswords := make([]string, len(records))

	for i, record := range records {
		// Generate a random temp password
		tempPassword, err := generateTempPassword(16) // 16 bytes = 24 chars in base64
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate password for row %d: %w", i+1, err)
		}

		// Hash the password
		hashedPassword, err := hashPassword(tempPassword)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to hash password for row %d: %w", i+1, err)
		}

		// Parse dates
		hireDate, err := time.Parse("2006-01-02", record.HireDate)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid hire_date for row %d: %w", i+1, err)
		}
		var dateOfBirth *time.Time
		if record.DateOfBirth != "" {
			dob, err := time.Parse("2006-01-02", record.DateOfBirth)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid date_of_birth for row %d: %w", i+1, err)
			}
			dateOfBirth = &dob
		}

		// Parse UUIDs
		departmentID, err := uuid.Parse(record.DepartmentID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid department_id for row %d: %w", i+1, err)
		}
		designationID, err := uuid.Parse(record.DesignationID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid designation_id for row %d: %w", i+1, err)
		}
		levelID, err := uuid.Parse(record.LevelID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid level_id for row %d: %w", i+1, err)
		}
		roleID, err := uuid.Parse(record.RoleID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid role_id for row %d: %w", i+1, err)
		}
		var managerID *uuid.UUID
		if record.ManagerID != "" {
			mid, err := uuid.Parse(record.ManagerID)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid manager_id for row %d: %w", i+1, err)
			}
			managerID = &mid
		}

		emp := &models.Employee{
			CompanyID:             companyID,
			Email:                 record.Email,
			PasswordHash:          hashedPassword,
			Phone:                 record.Phone,
			FirstName:             record.FirstName,
			LastName:              record.LastName,
			EmployeeCode:          record.EmployeeCode,
			DepartmentID:          &departmentID,
			DesignationID:         &designationID,
			LevelID:               &levelID,
			ManagerID:             managerID,
			RoleID:                roleID,
			Status:                record.Status,
			EmploymentType:        record.EmploymentType,
			HireDate:              hireDate,
			DateOfBirth:           dateOfBirth,
			Gender:                record.Gender,
			Address:               record.Address,
			EmergencyContactName:  record.EmergencyContactName,
			EmergencyContactPhone: record.EmergencyContactPhone,
			ProfileImageURL:       record.ProfileImageUrl,
		}

		employees[i] = emp
		tempPasswords[i] = tempPassword
	}

	return employees, tempPasswords, nil
}

// generateTempPassword creates a cryptographically secure random password
func generateTempPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
