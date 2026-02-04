package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type ILevelService interface {
	CreateLevel(ctx context.Context, companyID uuid.UUID, req *dto.CreateLevelRequest) (*dto.LevelResponse, error)
	GetLevelByID(ctx context.Context, levelID uuid.UUID) (*dto.LevelResponse, error)
	GetLevelList(ctx context.Context, companyID uuid.UUID, listRequest *dto.LevelListRequest) (*utils.PaginatedResponse[*models.Level], error)
	UpdateLevel(ctx context.Context, levelID uuid.UUID, req *dto.UpdateLevelRequest) (*dto.LevelResponse, error)
	DeleteLevel(ctx context.Context, levelID uuid.UUID) error
}

type LevelService struct {
	levelRepo *repositories.LevelRepository
}

func NewLevelService(levelRepo *repositories.LevelRepository) *LevelService {
	return &LevelService{levelRepo: levelRepo}
}

func (ls *LevelService) CreateLevel(
	ctx context.Context,
	companyID uuid.UUID,
	req *dto.CreateLevelRequest,
) (*dto.LevelResponse, error) {
	if req == nil {
		return nil, nil
	}

	level := &models.Level{
		CompanyID:      companyID,
		Name:           req.Name,
		HierarchyLevel: req.HierarchyLevel,
		MinSalary:      req.MinSalary,
		MaxSalary:      req.MaxSalary,
		Description:    req.Description,
	}

	created, err := ls.levelRepo.CreateLevel(ctx, level)
	if err != nil {
		return nil, err
	}

	return toLevelResponse(created), nil
}

func (ls *LevelService) GetLevelByID(ctx context.Context, levelID uuid.UUID) (*dto.LevelResponse, error) {
	level, err := ls.levelRepo.GetLevelByID(ctx, levelID)
	if err != nil {
		return nil, err
	}
	return toLevelResponse(level), nil
}

func (ls *LevelService) GetLevelList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.LevelListRequest,
) (*utils.PaginatedResponse[*models.Level], error) {
	return ls.levelRepo.GetLevelList(ctx, companyID, listRequest)
}

func (ls *LevelService) UpdateLevel(
	ctx context.Context,
	levelID uuid.UUID,
	req *dto.UpdateLevelRequest,
) (*dto.LevelResponse, error) {
	if req == nil {
		return nil, nil
	}

	level := &models.Level{}

	if req.Name != nil {
		level.Name = *req.Name
	}
	if req.HierarchyLevel != nil {
		level.HierarchyLevel = *req.HierarchyLevel
	}
	if req.MinSalary != nil {
		level.MinSalary = req.MinSalary
	}
	if req.MaxSalary != nil {
		level.MaxSalary = req.MaxSalary
	}
	if req.Description != nil {
		level.Description = *req.Description
	}

	updated, err := ls.levelRepo.UpdateLevel(ctx, levelID, level)
	if err != nil {
		return nil, err
	}

	return toLevelResponse(updated), nil
}

func (ls *LevelService) DeleteLevel(ctx context.Context, levelID uuid.UUID) error {
	return ls.levelRepo.DeleteLevel(ctx, levelID)
}

func toLevelResponse(level *models.Level) *dto.LevelResponse {
	if level == nil {
		return nil
	}

	return &dto.LevelResponse{
		ID:             level.ID.String(),
		CompanyID:      level.CompanyID.String(),
		Name:           level.Name,
		HierarchyLevel: level.HierarchyLevel,
		MinSalary:      level.MinSalary,
		MaxSalary:      level.MaxSalary,
		Description:    level.Description,
		CreatedAt:      level.CreatedAt,
		UpdatedAt:      level.UpdatedAt,
	}
}
