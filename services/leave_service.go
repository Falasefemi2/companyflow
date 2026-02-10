package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
	"github.com/google/uuid"
)

type ILeaveService interface {
	// Leave Type operations
	CreateLeaveType(ctx context.Context, companyID uuid.UUID, req *dto.CreateLeaveTypeRequest) (*models.LeaveType, error)
	GetLeaveTypeByID(ctx context.Context, leaveTypeID uuid.UUID) (*models.LeaveType, error)
	GetLeaveTypeList(ctx context.Context, companyID uuid.UUID, listRequest *dto.LeaveTypeListRequest) (*utils.PaginatedResponse[*models.LeaveType], error)
	UpdateLeaveType(ctx context.Context, leaveTypeID uuid.UUID, req *dto.UpdateLeaveTypeRequest) (*models.LeaveType, error)
	DeleteLeaveType(ctx context.Context, leaveTypeID uuid.UUID) error

	// Leave Request operations
	RequestLeave(ctx context.Context, employeeID uuid.UUID, req *dto.RequestLeaveRequest) (*models.LeaveRequest, error)
	GetLeaveRequest(ctx context.Context, requestID uuid.UUID) (*models.LeaveRequest, error)
	GetLeaveRequests(ctx context.Context, listRequest *dto.LeaveRequestListRequest) (*utils.PaginatedResponse[*models.LeaveRequest], error)
	ApproveLeaveRequest(ctx context.Context, requestID uuid.UUID, approvedByID uuid.UUID) (*models.LeaveRequest, error)
	RejectLeaveRequest(ctx context.Context, requestID uuid.UUID, req *dto.RejectLeaveRequestRequest) (*models.LeaveRequest, error)
	WithdrawLeaveRequest(ctx context.Context, requestID uuid.UUID, employeeID uuid.UUID) (*models.LeaveRequest, error)

	// Leave Balance operations
	CheckAvailableBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (float64, error)
	GetEmployeeLeaveBalances(ctx context.Context, employeeID uuid.UUID, year int) ([]*models.LeaveBalance, error)
	GetLeaveBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*models.LeaveBalance, error)
}

type LeaveService struct {
	repo repositories.ILeaveRepository
}

func NewLeaveService(repo repositories.ILeaveRepository) *LeaveService {
	return &LeaveService{
		repo: repo,
	}
}

func (s *LeaveService) CreateLeaveType(ctx context.Context, companyID uuid.UUID, req *dto.CreateLeaveTypeRequest) (*models.LeaveType, error) {
	if err := s.validateCreateLeaveTypeRequest(req); err != nil {
		return nil, err
	}

	leaveType := &models.LeaveType{
		CompanyID:             companyID,
		Name:                  req.Name,
		Code:                  req.Code,
		Description:           req.Description,
		DaysAllowed:           req.DaysAllowed,
		IsPaid:                req.IsPaid,
		RequiresDocumentation: req.RequiresDocumentation,
		CarryForwardAllowed:   req.CarryForwardAllowed,
		MaxCarryForwardDays:   req.MaxCarryForwardDays,
		ColorCode:             req.ColorCode,
		Status:                "active",
	}

	result, err := s.repo.CreateLeaveType(ctx, leaveType)
	if err != nil {
		return nil, fmt.Errorf("failed to create leave type: %w", err)
	}

	return result, nil
}

func (s *LeaveService) GetLeaveTypeByID(ctx context.Context, leaveTypeID uuid.UUID) (*models.LeaveType, error) {
	if leaveTypeID == uuid.Nil {
		return nil, errors.New("leave type ID cannot be empty")
	}

	result, err := s.repo.GetLeaveTypeByID(ctx, leaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("leave type not found: %w", err)
	}

	return result, nil
}

func (s *LeaveService) GetLeaveTypeList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.LeaveTypeListRequest,
) (*utils.PaginatedResponse[*models.LeaveType], error) {
	if companyID == uuid.Nil {
		return nil, errors.New("company ID cannot be empty")
	}

	if listRequest.Page < 1 {
		listRequest.Page = 1
	}
	if listRequest.PageSize < 1 || listRequest.PageSize > 100 {
		listRequest.PageSize = 10
	}

	result, err := s.repo.GetLeaveTypeList(ctx, companyID, listRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave types: %w", err)
	}

	return result, nil
}

func (s *LeaveService) UpdateLeaveType(ctx context.Context, leaveTypeID uuid.UUID, req *dto.UpdateLeaveTypeRequest) (*models.LeaveType, error) {
	if leaveTypeID == uuid.Nil {
		return nil, errors.New("leave type ID cannot be empty")
	}

	if err := s.validateUpdateLeaveTypeRequest(req); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetLeaveTypeByID(ctx, leaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("leave type not found: %w", err)
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Code != "" {
		existing.Code = req.Code
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.DaysAllowed > 0 {
		existing.DaysAllowed = req.DaysAllowed
	}
	if req.IsPaid != nil {
		existing.IsPaid = *req.IsPaid
	}
	if req.RequiresDocumentation != nil {
		existing.RequiresDocumentation = *req.RequiresDocumentation
	}
	if req.CarryForwardAllowed != nil {
		existing.CarryForwardAllowed = *req.CarryForwardAllowed
	}
	if req.MaxCarryForwardDays > 0 {
		existing.MaxCarryForwardDays = req.MaxCarryForwardDays
	}
	if req.ColorCode != "" {
		existing.ColorCode = req.ColorCode
	}
	if req.Status != "" {
		existing.Status = req.Status
	}

	result, err := s.repo.UpdateLeaveType(ctx, leaveTypeID, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update leave type: %w", err)
	}

	return result, nil
}

func (s *LeaveService) DeleteLeaveType(ctx context.Context, leaveTypeID uuid.UUID) error {
	if leaveTypeID == uuid.Nil {
		return errors.New("leave type ID cannot be empty")
	}

	err := s.repo.DeleteLeaveType(ctx, leaveTypeID)
	if err != nil {
		return fmt.Errorf("failed to delete leave type: %w", err)
	}

	return nil
}

func (s *LeaveService) RequestLeave(ctx context.Context, employeeID uuid.UUID, req *dto.RequestLeaveRequest) (*models.LeaveRequest, error) {
	if employeeID == uuid.Nil {
		return nil, errors.New("employee ID cannot be empty")
	}

	if err := s.validateRequestLeaveRequest(req); err != nil {
		return nil, err
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format. Use YYYY-MM-DD: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format. Use YYYY-MM-DD: %w", err)
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	today := time.Now().Truncate(24 * time.Hour)
	if startDate.Before(today) {
		return nil, errors.New("cannot request leave for past dates")
	}

	result, err := s.repo.CreateLeaveRequest(
		ctx,
		employeeID,
		req.LeaveTypeID,
		req.StartDate,
		req.EndDate,
		req.DaysRequested,
		req.Reason,
		&req.Attachment,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create leave request: %w", err)
	}

	return result, nil
}

func (s *LeaveService) GetLeaveRequest(ctx context.Context, requestID uuid.UUID) (*models.LeaveRequest, error) {
	if requestID == uuid.Nil {
		return nil, errors.New("request ID cannot be empty")
	}

	result, err := s.repo.GetLeaveRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("leave request not found: %w", err)
	}

	return result, nil
}

func (s *LeaveService) GetLeaveRequests(
	ctx context.Context,
	listRequest *dto.LeaveRequestListRequest,
) (*utils.PaginatedResponse[*models.LeaveRequest], error) {
	if listRequest.Page < 1 {
		listRequest.Page = 1
	}
	if listRequest.PageSize < 1 || listRequest.PageSize > 100 {
		listRequest.PageSize = 10
	}

	result, err := s.repo.GetLeaveRequestList(ctx, listRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave requests: %w", err)
	}

	return result, nil
}

func (s *LeaveService) ApproveLeaveRequest(ctx context.Context, requestID uuid.UUID, approvedByID uuid.UUID) (*models.LeaveRequest, error) {
	if requestID == uuid.Nil {
		return nil, errors.New("request ID cannot be empty")
	}
	if approvedByID == uuid.Nil {
		return nil, errors.New("approver ID cannot be empty")
	}

	request, err := s.repo.GetLeaveRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("leave request not found: %w", err)
	}

	if request.Status != "pending" {
		return nil, fmt.Errorf("can only approve pending requests. Current status: %s", request.Status)
	}

	result, err := s.repo.ApproveLeaveRequest(ctx, requestID, approvedByID)
	if err != nil {
		return nil, fmt.Errorf("failed to approve leave request: %w", err)
	}

	return result, nil
}

func (s *LeaveService) RejectLeaveRequest(ctx context.Context, requestID uuid.UUID, req *dto.RejectLeaveRequestRequest) (*models.LeaveRequest, error) {
	if requestID == uuid.Nil {
		return nil, errors.New("request ID cannot be empty")
	}

	if err := s.validateRejectLeaveRequest(req); err != nil {
		return nil, err
	}

	request, err := s.repo.GetLeaveRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("leave request not found: %w", err)
	}

	if request.Status != "pending" {
		return nil, fmt.Errorf("can only reject pending requests. Current status: %s", request.Status)
	}

	result, err := s.repo.RejectLeaveRequest(ctx, requestID, req.RejectionReason)
	if err != nil {
		return nil, fmt.Errorf("failed to reject leave request: %w", err)
	}

	return result, nil
}

func (s *LeaveService) WithdrawLeaveRequest(ctx context.Context, requestID uuid.UUID, employeeID uuid.UUID) (*models.LeaveRequest, error) {
	if requestID == uuid.Nil {
		return nil, errors.New("request ID cannot be empty")
	}
	if employeeID == uuid.Nil {
		return nil, errors.New("employee ID cannot be empty")
	}

	request, err := s.repo.GetLeaveRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("leave request not found: %w", err)
	}

	if request.EmployeeID != employeeID {
		return nil, errors.New("you can only withdraw your own leave requests")
	}

	if request.Status != "pending" {
		return nil, fmt.Errorf("can only withdraw pending requests. Current status: %s", request.Status)
	}

	result, err := s.repo.WithdrawLeaveRequest(ctx, requestID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to withdraw leave request: %w", err)
	}

	return result, nil
}

func (s *LeaveService) CheckAvailableBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (float64, error) {
	if employeeID == uuid.Nil {
		return 0, errors.New("employee ID cannot be empty")
	}
	if leaveTypeID == uuid.Nil {
		return 0, errors.New("leave type ID cannot be empty")
	}
	if year < 2020 || year > 2100 {
		return 0, errors.New("invalid year")
	}

	available, err := s.repo.CheckBalance(ctx, employeeID, leaveTypeID, year)
	if err != nil {
		return 0, fmt.Errorf("failed to check balance: %w", err)
	}

	return available, nil
}

func (s *LeaveService) GetEmployeeLeaveBalances(ctx context.Context, employeeID uuid.UUID, year int) ([]*models.LeaveBalance, error) {
	if employeeID == uuid.Nil {
		return nil, errors.New("employee ID cannot be empty")
	}
	if year < 2020 || year > 2100 {
		return nil, errors.New("invalid year")
	}

	result, err := s.repo.GetEmployeeBalances(ctx, employeeID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave balances: %w", err)
	}

	return result, nil
}

func (s *LeaveService) GetLeaveBalance(ctx context.Context, employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*models.LeaveBalance, error) {
	if employeeID == uuid.Nil {
		return nil, errors.New("employee ID cannot be empty")
	}
	if leaveTypeID == uuid.Nil {
		return nil, errors.New("leave type ID cannot be empty")
	}
	if year < 2020 || year > 2100 {
		return nil, errors.New("invalid year")
	}

	result, err := s.repo.GetLeaveBalance(ctx, employeeID, leaveTypeID, year)
	if err != nil {
		return nil, fmt.Errorf("leave balance not found: %w", err)
	}

	return result, nil
}

func (s *LeaveService) validateCreateLeaveTypeRequest(req *dto.CreateLeaveTypeRequest) error {
	if req.Name == "" {
		return errors.New("leave type name is required")
	}
	if len(req.Name) > 100 {
		return errors.New("leave type name must be less than 100 characters")
	}

	if req.Code == "" {
		return errors.New("leave type code is required")
	}
	if len(req.Code) > 50 {
		return errors.New("leave type code must be less than 50 characters")
	}

	if req.DaysAllowed < 0 {
		return errors.New("days allowed cannot be negative")
	}

	if req.ColorCode != "" && !isValidHexColor(req.ColorCode) {
		return errors.New("invalid color code format")
	}

	return nil
}

func (s *LeaveService) validateUpdateLeaveTypeRequest(req *dto.UpdateLeaveTypeRequest) error {
	if req.Name != "" && len(req.Name) > 100 {
		return errors.New("leave type name must be less than 100 characters")
	}

	if req.Code != "" && len(req.Code) > 50 {
		return errors.New("leave type code must be less than 50 characters")
	}

	if req.DaysAllowed < 0 {
		return errors.New("days allowed cannot be negative")
	}

	if req.ColorCode != "" && !isValidHexColor(req.ColorCode) {
		return errors.New("invalid color code format")
	}

	return nil
}

func (s *LeaveService) validateRequestLeaveRequest(req *dto.RequestLeaveRequest) error {
	if req.LeaveTypeID == uuid.Nil {
		return errors.New("leave type is required")
	}

	if req.StartDate == "" {
		return errors.New("start date is required")
	}

	if req.EndDate == "" {
		return errors.New("end date is required")
	}

	if req.DaysRequested <= 0 {
		return errors.New("days requested must be greater than 0")
	}

	if req.Reason == "" {
		return errors.New("reason is required")
	}
	if len(req.Reason) > 500 {
		return errors.New("reason must be less than 500 characters")
	}

	return nil
}

func (s *LeaveService) validateRejectLeaveRequest(req *dto.RejectLeaveRequestRequest) error {
	if req.RejectionReason == "" {
		return errors.New("rejection reason is required")
	}
	if len(req.RejectionReason) > 500 {
		return errors.New("rejection reason must be less than 500 characters")
	}

	return nil
}

func isValidHexColor(color string) bool {
	if len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < len(color); i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
