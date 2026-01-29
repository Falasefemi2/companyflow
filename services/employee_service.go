package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IEmployeeService interface {
	CreateEmployee(ctx context.Context, req *dto.CreateEmployeeRequest) (*dto.EmployeeResponse, error)
	GetEmployeeByID(ctx context.Context, employeeeID uuid.UUID) (*dto.EmployeeResponse, error)
	GetEmployeeList(ctx context.Context, companyID uuid.UUID, listRequest *dto.EmployeeListRequest) (*utils.PaginatedResponse[*models.Employee], error)
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

func (es *EmployeeService) CreateEmployee(ctx context.Context, req *dto.CreateEmployeeRequest) (*dto.EmployeeResponse, error) {
	return nil, nil
}
