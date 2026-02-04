package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IEmployeeService interface {
	CreateEmployee(ctx context.Context, companyID uuid.UUID, requesterRole string, req *dto.CreateEmployeeRequest) (*dto.EmployeeResponse, error)
	GetEmployeeByID(ctx context.Context, employeeeID uuid.UUID) (*dto.EmployeeResponse, error)
	GetEmployeeList(ctx context.Context, companyID uuid.UUID, listRequest *dto.EmployeeListRequest) (*utils.PaginatedResponse[*models.Employee], error)
	UpdateEmployee(ctx context.Context, employeeID uuid.UUID, req *dto.UpdateEmployeeRequest) (*dto.EmployeeResponse, error)
	DeleteEmployee(ctx context.Context, employeeID string, hardDelete bool) error
}

type EmployeeService struct {
	employeeRepo *repositories.EmployeeRepository
}

func NewEmployeeService(employeeRepo *repositories.EmployeeRepository) *EmployeeService {
	return &EmployeeService{
		employeeRepo: employeeRepo,
	}
}

func (es *EmployeeService) CreateEmployee(
	ctx context.Context,
	companyID uuid.UUID,
	requesterRole string,
	req *dto.CreateEmployeeRequest,
) (*dto.EmployeeResponse, error) {
	if req == nil {
		return nil, nil
	}

	if requesterRole != "Super Admin" && requesterRole != "HR Manager" {
		return nil, &utils.ValidationError{
			Field:   "role",
			Message: "insufficient role",
		}
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, &utils.ValidationError{
			Field:   "role_id",
			Message: "invalid role_id",
		}
	}

	_, roleCompanyID, err := es.employeeRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if roleCompanyID != nil && roleCompanyID.String() != companyID.String() {
		return nil, &utils.ValidationError{
			Field:   "role_id",
			Message: "role_id does not belong to company",
		}
	}

	departmentID, err := uuid.Parse(req.DepartmentID)
	if err != nil {
		return nil, &utils.ValidationError{
			Field:   "department_id",
			Message: "invalid department_id",
		}
	}

	designationID, err := uuid.Parse(req.DesignationID)
	if err != nil {
		return nil, &utils.ValidationError{
			Field:   "designation_id",
			Message: "invalid designation_id",
		}
	}

	levelID, err := uuid.Parse(req.LevelID)
	if err != nil {
		return nil, &utils.ValidationError{
			Field:   "level_id",
			Message: "invalid level_id",
		}
	}

	var managerID *uuid.UUID
	if req.ManagerID != "" {
		parsed, err := uuid.Parse(req.ManagerID)
		if err != nil {
			return nil, &utils.ValidationError{
				Field:   "manager_id",
				Message: "invalid manager_id",
			}
		}
		managerID = &parsed
	}

	dateOfBirth, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, &utils.ValidationError{
			Field:   "date_of_birth",
			Message: "invalid date_of_birth",
		}
	}

	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return nil, &utils.ValidationError{
			Field:   "hire_date",
			Message: "invalid hire_date",
		}
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	employee := &models.Employee{
		CompanyID:             companyID,
		Email:                 req.Email,
		PasswordHash:          hashedPassword,
		Phone:                 req.Phone,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		EmployeeCode:          req.EmployeeCode,
		DepartmentID:          &departmentID,
		DesignationID:         &designationID,
		LevelID:               &levelID,
		ManagerID:             managerID,
		RoleID:                roleID,
		Status:                req.Status,
		EmploymentType:        req.EmploymentType,
		HireDate:              hireDate,
		DateOfBirth:           &dateOfBirth,
		Gender:                req.Gender,
		Address:               req.Address,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		ProfileImageURL:       req.ProfileImageUrl,
	}

	created, err := es.employeeRepo.CreateEmployee(ctx, employee)
	if err != nil {
		return nil, err
	}

	return toEmployeeResponse(created), nil
}

func (es *EmployeeService) GetEmployeeByID(ctx context.Context, employeeID uuid.UUID) (*dto.EmployeeResponse, error) {
	employee, err := es.employeeRepo.GetEmployeeByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return toEmployeeResponse(employee), nil
}

func (es *EmployeeService) GetEmployeeList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.EmployeeListRequest,
) (*utils.PaginatedResponse[*models.Employee], error) {
	return es.employeeRepo.GetEmployeeList(ctx, companyID, listRequest)
}

func (es *EmployeeService) UpdateEmployee(
	ctx context.Context,
	employeeID uuid.UUID,
	req *dto.UpdateEmployeeRequest,
) (*dto.EmployeeResponse, error) {
	if req == nil {
		return nil, nil
	}

	employee := &models.Employee{}

	if req.Phone != nil {
		employee.Phone = *req.Phone
	}
	if req.FirstName != nil {
		employee.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		employee.LastName = *req.LastName
	}
	if req.DateOfBirth != nil {
		if *req.DateOfBirth == "" {
			employee.DateOfBirth = nil
		} else {
			parsed, err := time.Parse("2006-01-02", *req.DateOfBirth)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "date_of_birth",
					Message: "invalid date_of_birth",
				}
			}
			employee.DateOfBirth = &parsed
		}
	}
	if req.DepartmentID != nil {
		if *req.DepartmentID == "" {
			employee.DepartmentID = nil
		} else {
			parsed, err := uuid.Parse(*req.DepartmentID)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "department_id",
					Message: "invalid department_id",
				}
			}
			employee.DepartmentID = &parsed
		}
	}
	if req.DesignationID != nil {
		if *req.DesignationID == "" {
			employee.DesignationID = nil
		} else {
			parsed, err := uuid.Parse(*req.DesignationID)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "designation_id",
					Message: "invalid designation_id",
				}
			}
			employee.DesignationID = &parsed
		}
	}
	if req.LevelID != nil {
		if *req.LevelID == "" {
			employee.LevelID = nil
		} else {
			parsed, err := uuid.Parse(*req.LevelID)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "level_id",
					Message: "invalid level_id",
				}
			}
			employee.LevelID = &parsed
		}
	}
	if req.ManagerID != nil {
		if *req.ManagerID == "" {
			employee.ManagerID = nil
		} else {
			parsed, err := uuid.Parse(*req.ManagerID)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "manager_id",
					Message: "invalid manager_id",
				}
			}
			employee.ManagerID = &parsed
		}
	}
	if req.Status != nil {
		employee.Status = *req.Status
	}
	if req.Gender != nil {
		employee.Gender = *req.Gender
	}
	if req.Address != nil {
		employee.Address = *req.Address
	}
	if req.EmergencyContactName != nil {
		employee.EmergencyContactName = *req.EmergencyContactName
	}
	if req.EmergencyContactPhone != nil {
		employee.EmergencyContactPhone = *req.EmergencyContactPhone
	}
	if req.ProfileImageUrl != nil {
		employee.ProfileImageURL = *req.ProfileImageUrl
	}
	if req.TerminationDate != nil {
		if *req.TerminationDate == "" {
			employee.TerminationDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", *req.TerminationDate)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "termination_date",
					Message: "invalid termination_date",
				}
			}
			employee.TerminationDate = &parsed
		}
	}

	updated, err := es.employeeRepo.UpdateEmployee(ctx, employeeID, employee)
	if err != nil {
		return nil, err
	}

	return toEmployeeResponse(updated), nil
}

func (es *EmployeeService) DeleteEmployee(ctx context.Context, employeeID string, hardDelete bool) error {
	return es.employeeRepo.DeleteEmployee(ctx, employeeID, hardDelete)
}

func toEmployeeResponse(employee *models.Employee) *dto.EmployeeResponse {
	if employee == nil {
		return nil
	}

	var departmentID string
	if employee.DepartmentID != nil {
		departmentID = employee.DepartmentID.String()
	}

	var designationID string
	if employee.DesignationID != nil {
		designationID = employee.DesignationID.String()
	}

	var levelID string
	if employee.LevelID != nil {
		levelID = employee.LevelID.String()
	}

	var managerID string
	if employee.ManagerID != nil {
		managerID = employee.ManagerID.String()
	}

	return &dto.EmployeeResponse{
		ID:                    employee.ID.String(),
		CompanyID:             employee.CompanyID.String(),
		Email:                 employee.Email,
		Phone:                 employee.Phone,
		FirstName:             employee.FirstName,
		LastName:              employee.LastName,
		EmployeeCode:          employee.EmployeeCode,
		DepartmentID:          departmentID,
		DesignationID:         designationID,
		LevelID:               levelID,
		ManagerID:             managerID,
		RoleID:                employee.RoleID.String(),
		Status:                employee.Status,
		EmploymentType:        employee.EmploymentType,
		DateOfBirth:           employee.DateOfBirth,
		HireDate:              employee.HireDate,
		TerminationDate:       employee.TerminationDate,
		Gender:                employee.Gender,
		Address:               employee.Address,
		EmergencyContactName:  employee.EmergencyContactName,
		EmergencyContactPhone: employee.EmergencyContactPhone,
		ProfileImageUrl:       employee.ProfileImageURL,
		LastLoginAt:           employee.LastLoginAt,
		CreatedAt:             employee.CreatedAt,
		UpdatedAt:             employee.UpdatedAt,
	}
}
