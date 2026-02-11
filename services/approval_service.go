package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
)

type IApprovalService interface {
	CreateWorkflow(ctx context.Context, req *dto.CreateApprovalWorkflowRequest) (*models.ApprovalWorkflow, error)
	GetWorkflowList(ctx context.Context, req *dto.ApprovalWorkflowListRequest) ([]*models.ApprovalWorkflow, error)
	GetApprovalHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.ApprovalHistory, error)
}

type ApprovalService struct {
	repo repositories.IApprovalRepository
}

func NewApprovalService(repo repositories.IApprovalRepository) *ApprovalService {
	return &ApprovalService{repo: repo}
}

func (s *ApprovalService) CreateWorkflow(ctx context.Context, req *dto.CreateApprovalWorkflowRequest) (*models.ApprovalWorkflow, error) {
	if req.CompanyID == uuid.Nil {
		return nil, errors.New("company id is required")
	}
	if req.WorkflowType == "" {
		return nil, errors.New("workflow type is required")
	}
	if len(req.Steps) == 0 || !json.Valid(req.Steps) {
		return nil, errors.New("workflow steps must be a valid JSON array")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	workflow := &models.ApprovalWorkflow{
		CompanyID:    req.CompanyID,
		WorkflowType: req.WorkflowType,
		DepartmentID: req.DepartmentID,
		Steps:        req.Steps,
		IsActive:     isActive,
	}

	result, err := s.repo.CreateWorkflow(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval workflow: %w", err)
	}
	return result, nil
}

func (s *ApprovalService) GetWorkflowList(ctx context.Context, req *dto.ApprovalWorkflowListRequest) ([]*models.ApprovalWorkflow, error) {
	if req.CompanyID == uuid.Nil {
		return nil, errors.New("company id is required")
	}
	result, err := s.repo.GetWorkflowList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list approval workflows: %w", err)
	}
	return result, nil
}

func (s *ApprovalService) GetApprovalHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.ApprovalHistory, error) {
	if entityType == "" {
		return nil, errors.New("entity type is required")
	}
	if entityID == uuid.Nil {
		return nil, errors.New("entity id is required")
	}

	result, err := s.repo.GetApprovalHistory(ctx, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to load approval history: %w", err)
	}
	return result, nil
}
