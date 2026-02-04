package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IDesignationService interface {
	CreateDesignation(ctx context.Context, companyID uuid.UUID, req *dto.CreateDesignationRequest) (*dto.DesignationResponse, error)
	GetDesignationByID(ctx context.Context, designationID uuid.UUID) (*dto.DesignationResponse, error)
	GetDesignationList(ctx context.Context, companyID uuid.UUID, listRequest *dto.DesignationListRequest) (*utils.PaginatedResponse[*models.Designation], error)
	UpdateDesignation(ctx context.Context, designationID uuid.UUID, req *dto.UpdateDesignationRequest) (*dto.DesignationResponse, error)
	DeleteDesignation(ctx context.Context, designationID uuid.UUID) error
}

type DesignationService struct {
	designationRepo *repositories.DesignationRepository
}

func NewDesignationService(designationRepo *repositories.DesignationRepository) *DesignationService {
	return &DesignationService{designationRepo: designationRepo}
}

func (ds *DesignationService) CreateDesignation(
	ctx context.Context,
	companyID uuid.UUID,
	req *dto.CreateDesignationRequest,
) (*dto.DesignationResponse, error) {
	if req == nil {
		return nil, nil
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	var levelID *uuid.UUID
	if req.LevelID != "" {
		parsed, err := uuid.Parse(req.LevelID)
		if err != nil {
			return nil, &utils.ValidationError{
				Field:   "level_id",
				Message: "invalid level_id",
			}
		}
		levelID = &parsed
	}

	var departmentID *uuid.UUID
	if req.DepartmentID != "" {
		parsed, err := uuid.Parse(req.DepartmentID)
		if err != nil {
			return nil, &utils.ValidationError{
				Field:   "department_id",
				Message: "invalid department_id",
			}
		}
		departmentID = &parsed
	}

	designation := &models.Designation{
		CompanyID:    companyID,
		Name:         req.Name,
		Description:  req.Description,
		LevelID:      levelID,
		DepartmentID: departmentID,
		Status:       status,
	}

	created, err := ds.designationRepo.CreateDesignation(ctx, designation)
	if err != nil {
		return nil, err
	}

	return toDesignationResponse(created), nil
}

func (ds *DesignationService) GetDesignationByID(ctx context.Context, designationID uuid.UUID) (*dto.DesignationResponse, error) {
	designation, err := ds.designationRepo.GetDesignationByID(ctx, designationID)
	if err != nil {
		return nil, err
	}
	return toDesignationResponse(designation), nil
}

func (ds *DesignationService) GetDesignationList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.DesignationListRequest,
) (*utils.PaginatedResponse[*models.Designation], error) {
	return ds.designationRepo.GetDesignationList(ctx, companyID, listRequest)
}

func (ds *DesignationService) UpdateDesignation(
	ctx context.Context,
	designationID uuid.UUID,
	req *dto.UpdateDesignationRequest,
) (*dto.DesignationResponse, error) {
	if req == nil {
		return nil, nil
	}

	designation := &models.Designation{}

	if req.Name != nil {
		designation.Name = *req.Name
	}
	if req.Description != nil {
		designation.Description = *req.Description
	}
	if req.LevelID != nil {
		if *req.LevelID == "" {
			designation.LevelID = nil
		} else {
			parsed, err := uuid.Parse(*req.LevelID)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "level_id",
					Message: "invalid level_id",
				}
			}
			designation.LevelID = &parsed
		}
	}
	if req.DepartmentID != nil {
		if *req.DepartmentID == "" {
			designation.DepartmentID = nil
		} else {
			parsed, err := uuid.Parse(*req.DepartmentID)
			if err != nil {
				return nil, &utils.ValidationError{
					Field:   "department_id",
					Message: "invalid department_id",
				}
			}
			designation.DepartmentID = &parsed
		}
	}
	if req.Status != nil {
		designation.Status = *req.Status
	}

	updated, err := ds.designationRepo.UpdateDesignation(ctx, designationID, designation)
	if err != nil {
		return nil, err
	}

	return toDesignationResponse(updated), nil
}

func (ds *DesignationService) DeleteDesignation(ctx context.Context, designationID uuid.UUID) error {
	return ds.designationRepo.DeleteDesignation(ctx, designationID)
}

func toDesignationResponse(designation *models.Designation) *dto.DesignationResponse {
	if designation == nil {
		return nil
	}

	var levelID *string
	if designation.LevelID != nil {
		value := designation.LevelID.String()
		levelID = &value
	}

	var departmentID *string
	if designation.DepartmentID != nil {
		value := designation.DepartmentID.String()
		departmentID = &value
	}

	return &dto.DesignationResponse{
		ID:           designation.ID.String(),
		CompanyID:    designation.CompanyID.String(),
		Name:         designation.Name,
		Description:  designation.Description,
		LevelID:      levelID,
		DepartmentID: departmentID,
		Status:       designation.Status,
		CreatedAt:    designation.CreatedAt,
		UpdatedAt:    designation.UpdatedAt,
	}
}
