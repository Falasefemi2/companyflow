package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IDepartmentService interface {
	CreateDepartment(ctx context.Context, companyID uuid.UUID, req *dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error)
	GetDepartmentByID(ctx context.Context, departmentID uuid.UUID) (*dto.DepartmentResponse, error)
	GetDepartmentList(ctx context.Context, companyID uuid.UUID, listRequest *dto.DepartmentListRequest) (*utils.PaginatedResponse[*models.Department], error)
	UpdateDepartment(ctx context.Context, departmentID uuid.UUID, req *dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error)
	DeleteDepartment(ctx context.Context, departmentID uuid.UUID, softDelete bool) error
}

type DepartmentService struct {
	departmentRepo *repositories.DepartmentRepository
}

func NewDepartmentService(departmentRepo *repositories.DepartmentRepository) *DepartmentService {
	return &DepartmentService{departmentRepo: departmentRepo}
}

func (ds *DepartmentService) CreateDepartment(
	ctx context.Context,
	companyID uuid.UUID,
	req *dto.CreateDepartmentRequest,
) (*dto.DepartmentResponse, error) {
	if req == nil {
		return nil, nil
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	var parentID *uuid.UUID
	if req.ParentDepartmentID != "" {
		parsed, err := uuid.Parse(req.ParentDepartmentID)
		if err != nil {
			return nil, &utils.ValidationError{
				Field:   "parent_department_id",
				Message: "invalid parent_department_id",
			}
		}
		parentID = &parsed
	}

	department := &models.Department{
		CompanyID:          companyID,
		Name:               req.Name,
		Code:               req.Code,
		Description:        req.Description,
		ParentDepartmentID: parentID,
		CostCenter:         req.CostCenter,
		Status:             status,
	}

	created, err := ds.departmentRepo.CreateDepartment(ctx, department)
	if err != nil {
		return nil, err
	}

	return toDepartmentResponse(created), nil
}

func (ds *DepartmentService) GetDepartmentByID(ctx context.Context, departmentID uuid.UUID) (*dto.DepartmentResponse, error) {
	department, err := ds.departmentRepo.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	return toDepartmentResponse(department), nil
}

func (ds *DepartmentService) GetDepartmentList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.DepartmentListRequest,
) (*utils.PaginatedResponse[*models.Department], error) {
	return ds.departmentRepo.GetDepartmentList(ctx, companyID, listRequest)
}

func (ds *DepartmentService) UpdateDepartment(
	ctx context.Context,
	departmentID uuid.UUID,
	req *dto.UpdateDepartmentRequest,
) (*dto.DepartmentResponse, error) {
	if req == nil {
		return nil, nil
	}

	department := &models.Department{}

	if req.Name != nil {
		department.Name = *req.Name
	}
	if req.Code != nil {
		department.Code = *req.Code
	}
	if req.Description != nil {
		department.Description = *req.Description
	}
	if req.ParentDepartmentID != nil {
		if *req.ParentDepartmentID == "" {
			department.ParentDepartmentID = nil
		} else {
			parsed, err := uuid.Parse(*req.ParentDepartmentID)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "parent_department_id",
					Message: "invalid parent_department_id",
				}
			}
			department.ParentDepartmentID = &parsed
		}
	}
	if req.CostCenter != nil {
		department.CostCenter = *req.CostCenter
	}
	if req.Status != nil {
		department.Status = *req.Status
	}

	updated, err := ds.departmentRepo.UpdateDepartment(ctx, departmentID, department)
	if err != nil {
		return nil, err
	}

	return toDepartmentResponse(updated), nil
}

func (ds *DepartmentService) DeleteDepartment(ctx context.Context, departmentID uuid.UUID, softDelete bool) error {
	return ds.departmentRepo.DeleteDepartment(ctx, departmentID, softDelete)
}

func toDepartmentResponse(department *models.Department) *dto.DepartmentResponse {
	if department == nil {
		return nil
	}

	var parentID *string
	if department.ParentDepartmentID != nil {
		value := department.ParentDepartmentID.String()
		parentID = &value
	}

	return &dto.DepartmentResponse{
		ID:                 department.ID.String(),
		CompanyID:          department.CompanyID.String(),
		Name:               department.Name,
		Code:               department.Code,
		Description:        department.Description,
		ParentDepartmentID: parentID,
		CostCenter:         department.CostCenter,
		Status:             department.Status,
		CreatedAt:          department.CreatedAt,
		UpdatedAt:          department.UpdatedAt,
	}
}
