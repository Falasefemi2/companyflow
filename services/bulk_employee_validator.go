package services

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type BulkEmployeeValidator struct {
	employeeRepo    *repositories.EmployeeRepository
	departmentRepo  *repositories.DepartmentRepository
	roleRepo        *repositories.RoleRepository
	designationRepo *repositories.DesignationRepository
	levelRepo       *repositories.LevelRepository
}

func NewBulkEmployeeValidator(
	employeeRepo *repositories.EmployeeRepository,
	departmentRepo *repositories.DepartmentRepository,
	roleRepo *repositories.RoleRepository,
	designationRepo *repositories.DesignationRepository,
	levelRepo *repositories.LevelRepository,
) *BulkEmployeeValidator {
	return &BulkEmployeeValidator{
		employeeRepo:    employeeRepo,
		departmentRepo:  departmentRepo,
		roleRepo:        roleRepo,
		designationRepo: designationRepo,
		levelRepo:       levelRepo,
	}
}

func (v *BulkEmployeeValidator) ValidateBulkRecords(
	ctx context.Context,
	companyID uuid.UUID,
	records []dto.BulkEmployeeRecord,
) ([]dto.BulkEmployeeError, error) {
	// Load all reference data ONCE
	departmentMap, err := v.loadDepartments(ctx, companyID)
	if err != nil {
		return nil, err
	}

	roleMap, err := v.loadRoles(ctx, companyID)
	if err != nil {
		return nil, err
	}

	designationMap, err := v.loadDesignations(ctx, companyID)
	if err != nil {
		return nil, err
	}

	levelMap, err := v.loadLevels(ctx, companyID)
	if err != nil {
		return nil, err
	}

	emailMap, employeeCodeMap, err := v.loadEmployees(ctx, companyID)
	if err != nil {
		return nil, err
	}

	var bulkErrors []dto.BulkEmployeeError
	seenEmails := make(map[string]int)
	seenEmployeeCodes := make(map[string]int)

	for rowNum, record := range records {
		// Pass maps to validation
		rowErrors := v.validateSingleRecord(
			ctx,
			companyID,
			record,
			rowNum+2,
			departmentMap,
			roleMap,
			designationMap,
			levelMap,
			emailMap,
			employeeCodeMap,
			seenEmails,
			seenEmployeeCodes,
		)

		if len(rowErrors) > 0 {
			bulkErrors = append(bulkErrors, dto.BulkEmployeeError{
				RowNumber: rowNum + 2,
				Record:    record,
				Errors:    rowErrors,
			})
		}
	}

	return bulkErrors, nil
}

func (v *BulkEmployeeValidator) validateSingleRecord(
	ctx context.Context,
	companyID uuid.UUID,
	record dto.BulkEmployeeRecord,
	rowNumber int,
	departmentMap map[uuid.UUID]bool,
	roleMap map[uuid.UUID]bool,
	designationMap map[uuid.UUID]bool,
	levelMap map[uuid.UUID]bool,
	emailMap map[string]bool,
	employeeCodeMap map[string]bool,
	seenEmails map[string]int,
	seenEmployeeCodes map[string]int,
) []string {
	var errors []string
	normalizedEmail := normalizeEmail(record.Email)
	normalizedEmployeeCode := strings.TrimSpace(record.EmployeeCode)

	// Validate required fields and formats
	if record.Email == "" {
		errors = append(errors, "email is required")
	} else if !isValidEmail(record.Email) {
		errors = append(errors, "email format is invalid")
	} else if prevRow, exists := seenEmails[normalizedEmail]; exists {
		errors = append(errors, fmt.Sprintf("email %s is duplicated in CSV (rows %d and %d)", record.Email, prevRow, rowNumber))
	} else {
		seenEmails[normalizedEmail] = rowNumber
	}

	if record.FirstName == "" {
		errors = append(errors, "first_name is required")
	}

	if record.LastName == "" {
		errors = append(errors, "last_name is required")
	}

	if record.Phone == "" {
		errors = append(errors, "phone is required")
	}

	if record.EmployeeCode == "" {
		errors = append(errors, "employee_code is required")
	} else if prevRow, exists := seenEmployeeCodes[normalizedEmployeeCode]; exists {
		errors = append(errors, fmt.Sprintf("employee_code %s is duplicated in CSV (rows %d and %d)", record.EmployeeCode, prevRow, rowNumber))
	} else {
		seenEmployeeCodes[normalizedEmployeeCode] = rowNumber
	}

	if record.DateOfBirth != "" {
		if !isValidDate(record.DateOfBirth) {
			errors = append(errors, "date_of_birth must be in YYYY-MM-DD format")
		}
	}

	if record.HireDate == "" {
		errors = append(errors, "hire_date is required")
	} else if !isValidDate(record.HireDate) {
		errors = append(errors, "hire_date must be in YYYY-MM-DD format")
	}

	if record.DepartmentID == "" {
		errors = append(errors, "department_id is required")
	} else if !isValidUUID(record.DepartmentID) {
		errors = append(errors, "department_id must be a valid UUID")
	}

	if record.DesignationID == "" {
		errors = append(errors, "designation_id is required")
	} else if !isValidUUID(record.DesignationID) {
		errors = append(errors, "designation_id must be a valid UUID")
	}

	if record.LevelID == "" {
		errors = append(errors, "level_id is required")
	} else if !isValidUUID(record.LevelID) {
		errors = append(errors, "level_id must be a valid UUID")
	}

	if record.RoleID == "" {
		errors = append(errors, "role_id is required")
	} else if !isValidUUID(record.RoleID) {
		errors = append(errors, "role_id must be a valid UUID")
	}

	if record.ManagerID != "" && !isValidUUID(record.ManagerID) {
		errors = append(errors, "manager_id must be a valid UUID")
	}

	if record.Status == "" {
		errors = append(errors, "status is required")
	} else if !isValidStatus(record.Status) {
		errors = append(errors, fmt.Sprintf("status must be one of: active, inactive, on_leave, terminated, probation. Got: %s", record.Status))
	}

	if record.EmploymentType == "" {
		errors = append(errors, "employment_type is required")
	} else if !isValidEmploymentType(record.EmploymentType) {
		errors = append(errors, fmt.Sprintf("employment_type must be one of: full_time, part_time, contract, intern. Got: %s", record.EmploymentType))
	}

	// Only check foreign keys if basic validation passed
	if len(errors) == 0 {
		fkErrors := v.validateForeignKeys(
			ctx,
			companyID,
			record,
			departmentMap,
			roleMap,
			designationMap,
			levelMap,
			emailMap,
			employeeCodeMap,
			normalizedEmail,
			normalizedEmployeeCode,
		)
		errors = append(errors, fkErrors...)
	}

	return errors
}

func (v *BulkEmployeeValidator) validateForeignKeys(
	ctx context.Context,
	companyID uuid.UUID,
	record dto.BulkEmployeeRecord,
	departmentMap map[uuid.UUID]bool,
	roleMap map[uuid.UUID]bool,
	designationMap map[uuid.UUID]bool,
	levelMap map[uuid.UUID]bool,
	emailMap map[string]bool,
	employeeCodeMap map[string]bool,
	normalizedEmail string,
	normalizedEmployeeCode string,
) []string {
	var errors []string

	// Check department exists
	deptID, err := uuid.Parse(record.DepartmentID)
	if err == nil && !departmentMap[deptID] {
		errors = append(errors, fmt.Sprintf("department_id %s not found in system", record.DepartmentID))
	}

	// Check role exists
	roleID, err := uuid.Parse(record.RoleID)
	if err == nil && !roleMap[roleID] {
		errors = append(errors, fmt.Sprintf("role_id %s not found in system", record.RoleID))
	}

	// Check designation exists
	designationID, err := uuid.Parse(record.DesignationID)
	if err == nil && !designationMap[designationID] {
		errors = append(errors, fmt.Sprintf("designation_id %s not found in system", record.DesignationID))
	}

	// Check level exists
	levelID, err := uuid.Parse(record.LevelID)
	if err == nil && !levelMap[levelID] {
		errors = append(errors, fmt.Sprintf("level_id %s not found in system", record.LevelID))
	}

	// Check manager exists (only if provided)
	if record.ManagerID != "" {
		managerID, err := uuid.Parse(record.ManagerID)
		if err == nil {
			// For now, we skip manager validation since we don't have a map of employee IDs
			// You could add this later if needed
			_ = managerID
		}
	}

	// Check email doesn't already exist
	if normalizedEmail != "" && emailMap[normalizedEmail] {
		errors = append(errors, fmt.Sprintf("email %s already exists in system", record.Email))
	}

	if normalizedEmployeeCode != "" && employeeCodeMap[normalizedEmployeeCode] {
		errors = append(errors, fmt.Sprintf("employee_code %s already exists in system", record.EmployeeCode))
	}

	return errors
}

// Helper validation functions
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func isValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

func isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"active":     true,
		"inactive":   true,
		"on_leave":   true,
		"terminated": true,
		"probation":  true,
	}
	return validStatuses[status]
}

func isValidEmploymentType(empType string) bool {
	validTypes := map[string]bool{
		"full_time": true,
		"part_time": true,
		"contract":  true,
		"intern":    true,
	}
	return validTypes[empType]
}

// Load reference data into maps for O(1) lookups
func (v *BulkEmployeeValidator) loadDepartments(ctx context.Context, companyID uuid.UUID) (map[uuid.UUID]bool, error) {
	deptMap := make(map[uuid.UUID]bool)
	page := 1
	for {
		listRequest := &dto.DepartmentListRequest{
			PaginationParams: utils.PaginationParams{
				Page:     page,
				PageSize: 100,
			},
		}
		response, err := v.departmentRepo.GetDepartmentList(ctx, companyID, listRequest)
		if err != nil {
			return nil, err
		}
		for _, dept := range response.Data {
			deptMap[dept.ID] = true
		}
		if !response.HasNext {
			break
		}
		page++
	}
	return deptMap, nil
}

func (v *BulkEmployeeValidator) loadRoles(ctx context.Context, companyID uuid.UUID) (map[uuid.UUID]bool, error) {
	roleMap := make(map[uuid.UUID]bool)
	page := 1
	for {
		listRequest := &dto.RoleListRequest{
			Page:     page,
			PageSize: 100,
		}

		response, err := v.roleRepo.GetRoleList(ctx, companyID, listRequest)
		if err != nil {
			return nil, err
		}
		for _, role := range response.Data {
			roleMap[role.ID] = true
		}
		if !response.HasNext {
			break
		}
		page++
	}
	return roleMap, nil
}

func (v *BulkEmployeeValidator) loadDesignations(ctx context.Context, companyID uuid.UUID) (map[uuid.UUID]bool, error) {
	designationMap := make(map[uuid.UUID]bool)
	page := 1
	for {
		listRequest := &dto.DesignationListRequest{
			PaginationParams: utils.PaginationParams{
				Page:     page,
				PageSize: 100,
			},
		}
		response, err := v.designationRepo.GetDesignationList(ctx, companyID, listRequest)
		if err != nil {
			return nil, err
		}
		for _, designation := range response.Data {
			designationMap[designation.ID] = true
		}
		if !response.HasNext {
			break
		}
		page++
	}
	return designationMap, nil
}

func (v *BulkEmployeeValidator) loadLevels(ctx context.Context, companyID uuid.UUID) (map[uuid.UUID]bool, error) {
	levelMap := make(map[uuid.UUID]bool)
	page := 1
	for {
		listRequest := &dto.LevelListRequest{
			PaginationParams: utils.PaginationParams{
				Page:     page,
				PageSize: 100,
			},
		}
		response, err := v.levelRepo.GetLevelList(ctx, companyID, listRequest)
		if err != nil {
			return nil, err
		}
		for _, level := range response.Data {
			levelMap[level.ID] = true
		}
		if !response.HasNext {
			break
		}
		page++
	}
	return levelMap, nil
}

func (v *BulkEmployeeValidator) loadEmployees(ctx context.Context, companyID uuid.UUID) (map[string]bool, map[string]bool, error) {
	emailMap := make(map[string]bool)
	employeeCodeMap := make(map[string]bool)
	page := 1
	for {
		listRequest := &dto.EmployeeListRequest{
			PaginationParams: utils.PaginationParams{
				Page:     page,
				PageSize: 100,
			},
		}
		response, err := v.employeeRepo.GetEmployeeList(ctx, companyID, listRequest)
		if err != nil {
			return nil, nil, err
		}
		for _, employee := range response.Data {
			emailMap[normalizeEmail(employee.Email)] = true
			employeeCodeMap[strings.TrimSpace(employee.EmployeeCode)] = true
		}
		if !response.HasNext {
			break
		}
		page++
	}
	return emailMap, employeeCodeMap, nil
}
