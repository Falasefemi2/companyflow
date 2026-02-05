package services

import (
	"context"
	"errors"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IAuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
}

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	employeeRepo *repositories.EmployeeRepository
	companyRepo  *repositories.CompanyRepository
}

func NewAuthService(
	employeeRepo *repositories.EmployeeRepository,
	companyRepo *repositories.CompanyRepository,
) *AuthService {
	return &AuthService{
		employeeRepo: employeeRepo,
		companyRepo:  companyRepo,
	}
}

func (as *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	if req == nil {
		return nil, ErrInvalidCredentials
	}

	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	employee, roleName, err := as.employeeRepo.GetEmployeeWithRoleByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if roleName != "Super Admin" && roleName != "HR Manager" {
		return nil, ErrInvalidCredentials
	}

	if employee.Status != "active" {
		return nil, ErrInvalidCredentials
	}

	if !utils.VerifyPassword(employee.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(employee.ID.String(), roleName, employee.CompanyID.String(), 24)
	if err != nil {
		return nil, err
	}

	company, err := as.companyRepo.GetCompanyByID(ctx, employee.CompanyID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:    token,
		Role:     roleName,
		Employee: ToEmployeeResponse(employee),
		Company:  ToCompanyResponse(company),
	}, nil
}
