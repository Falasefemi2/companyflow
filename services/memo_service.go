package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IMemoService interface {
	CreateMemo(ctx context.Context, employeeID uuid.UUID, req *dto.CreateMemoRequest) (*models.Memo, error)
	GetMemoByID(ctx context.Context, memoID uuid.UUID) (*models.Memo, error)
	GetMemos(ctx context.Context, req *dto.MemoListRequest) (*utils.PaginatedResponse[*models.Memo], error)
	MarkMemoRead(ctx context.Context, memoID uuid.UUID, employeeID uuid.UUID) error
	ApproveMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error)
	RejectMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error)
}

type MemoService struct {
	repo repositories.IMemoRepository
}

func NewMemoService(repo repositories.IMemoRepository) *MemoService {
	return &MemoService{repo: repo}
}

func (s *MemoService) CreateMemo(ctx context.Context, employeeID uuid.UUID, req *dto.CreateMemoRequest) (*models.Memo, error) {
	if employeeID == uuid.Nil {
		return nil, errors.New("employee id is required")
	}
	if req.CompanyID == uuid.Nil {
		return nil, errors.New("company id is required")
	}
	if req.MemoType == "" {
		return nil, errors.New("memo type is required")
	}
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.Content == "" {
		return nil, errors.New("content is required")
	}
	if req.Priority == "" {
		req.Priority = "normal"
	}

	memo := &models.Memo{
		CompanyID:       req.CompanyID,
		EmployeeID:      employeeID,
		MemoType:        req.MemoType,
		Title:           req.Title,
		Content:         req.Content,
		ReferenceNumber: req.ReferenceNumber,
		Status:          "pending",
		CurrentStep:     1,
		Priority:        req.Priority,
	}

	result, err := s.repo.CreateMemo(ctx, memo, req.RecipientIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create memo: %w", err)
	}
	return result, nil
}

func (s *MemoService) GetMemoByID(ctx context.Context, memoID uuid.UUID) (*models.Memo, error) {
	if memoID == uuid.Nil {
		return nil, errors.New("memo id is required")
	}
	result, err := s.repo.GetMemoByID(ctx, memoID)
	if err != nil {
		return nil, fmt.Errorf("memo not found: %w", err)
	}
	return result, nil
}

func (s *MemoService) GetMemos(ctx context.Context, req *dto.MemoListRequest) (*utils.PaginatedResponse[*models.Memo], error) {
	if req.CompanyID == uuid.Nil {
		return nil, errors.New("company id is required")
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	result, err := s.repo.GetMemoList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch memos: %w", err)
	}
	return result, nil
}

func (s *MemoService) MarkMemoRead(ctx context.Context, memoID uuid.UUID, employeeID uuid.UUID) error {
	if memoID == uuid.Nil {
		return errors.New("memo id is required")
	}
	if employeeID == uuid.Nil {
		return errors.New("employee id is required")
	}
	return s.repo.MarkMemoRead(ctx, memoID, employeeID)
}

func (s *MemoService) ApproveMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error) {
	if memoID == uuid.Nil {
		return nil, errors.New("memo id is required")
	}
	if approverID == uuid.Nil {
		return nil, errors.New("approver id is required")
	}
	return s.repo.ApproveMemo(ctx, memoID, approverID, comments)
}

func (s *MemoService) RejectMemo(ctx context.Context, memoID uuid.UUID, approverID uuid.UUID, comments string) (*models.Memo, error) {
	if memoID == uuid.Nil {
		return nil, errors.New("memo id is required")
	}
	if approverID == uuid.Nil {
		return nil, errors.New("approver id is required")
	}
	return s.repo.RejectMemo(ctx, memoID, approverID, comments)
}
