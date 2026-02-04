package services

import (
	"context"
	"errors"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IAuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest) (string, error)
}

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	employeeRepo *repositories.EmployeeRepository
}

func NewAuthService(employeeRepo *repositories.EmployeeRepository) *AuthService {
	return &AuthService{employeeRepo: employeeRepo}
}

func (as *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (string, error) {
	if req == nil {
		return "", ErrInvalidCredentials
	}

	if req.Email == "" || req.Password == "" {
		return "", ErrInvalidCredentials
	}

	employee, roleName, err := as.employeeRepo.GetEmployeeWithRoleByEmail(ctx, req.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if roleName != "Super Admin" && roleName != "HR Manager" {
		return "", ErrInvalidCredentials
	}

	if employee.Status != "active" {
		return "", ErrInvalidCredentials
	}

	if !utils.VerifyPassword(employee.PasswordHash, req.Password) {
		return "", ErrInvalidCredentials
	}

	return utils.GenerateToken(employee.ID.String(), roleName, employee.CompanyID.String(), 24)
}
