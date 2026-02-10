package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
)

func setupLeaveRepository(t *testing.T) *LeaveRepository {
	pool := setupTestDB(t)
	return NewLeaveRepository(pool)
}

func cleanupLeaveTestData(ctx context.Context, pool *pgxpool.Pool, companyID uuid.UUID) error {
	// Clean in order due to foreign key constraints
	queries := []string{
		"DELETE FROM leave_balances WHERE employee_id IN (SELECT id FROM employees WHERE company_id = $1)",
		"DELETE FROM leave_requests WHERE employee_id IN (SELECT id FROM employees WHERE company_id = $1)",
		"DELETE FROM leave_types WHERE company_id = $1",
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query, companyID); err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}
	}
	return nil
}

func createTestCompany(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	ctx := context.Background()
	companyID := uuid.New()

	query := `
		INSERT INTO companies (id, name, slug, status, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := pool.Exec(ctx, query, companyID, "Test Company", fmt.Sprintf("test-%d", time.Now().UnixNano()), "active", []byte(`{}`))
	if err != nil {
		t.Fatalf("failed to create test company: %v", err)
	}

	return companyID
}

func createTestEmployee(t *testing.T, pool *pgxpool.Pool, companyID uuid.UUID) uuid.UUID {
	ctx := context.Background()
	employeeID := uuid.New()

	query := `
		INSERT INTO employees (id, company_id, email, first_name, last_name, status, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	_, err := pool.Exec(ctx, query, employeeID, companyID, email, "Test", "Employee", "active", "hash")
	if err != nil {
		t.Fatalf("failed to create test employee: %v", err)
	}

	return employeeID
}

func TestLeaveRepository_CreateLeaveType(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)

	leaveType := &models.LeaveType{
		CompanyID:             companyID,
		Name:                  "Annual Leave",
		Code:                  "AL",
		Description:           "Annual vacation leave",
		DaysAllowed:           20.0,
		IsPaid:                true,
		RequiresDocumentation: false,
		CarryForwardAllowed:   true,
		MaxCarryForwardDays:   5.0,
		ColorCode:             "#4CAF50",
		Status:                "active",
	}

	result, err := repo.CreateLeaveType(ctx, leaveType)
	if err != nil {
		t.Fatalf("CreateLeaveType failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected leave type, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.Name != leaveType.Name {
		t.Errorf("expected name %s, got %s", leaveType.Name, result.Name)
	}

	if result.DaysAllowed != leaveType.DaysAllowed {
		t.Errorf("expected days allowed %.1f, got %.1f", leaveType.DaysAllowed, result.DaysAllowed)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestLeaveRepository_GetLeaveTypeByID(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)

	leaveType := &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Sick Leave",
		Code:        "SL",
		DaysAllowed: 10.0,
		IsPaid:      true,
		Status:      "active",
	}

	created, err := repo.CreateLeaveType(ctx, leaveType)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result, err := repo.GetLeaveTypeByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetLeaveTypeByID failed: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("expected ID %v, got %v", created.ID, result.ID)
	}

	if result.Name != created.Name {
		t.Errorf("expected name %s, got %s", created.Name, result.Name)
	}
}

func TestLeaveRepository_GetLeaveTypeByID_NotFound(t *testing.T) {
	repo := setupLeaveRepository(t)
	ctx := context.Background()

	nonExistentID := uuid.New()
	_, err := repo.GetLeaveTypeByID(ctx, nonExistentID)
	if err == nil {
		t.Error("expected error for non-existent leave type")
	}
}

func TestLeaveRepository_GetLeaveTypeList(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)

	// Create multiple leave types
	for i := 1; i <= 5; i++ {
		_, err := repo.CreateLeaveType(ctx, &models.LeaveType{
			CompanyID:   companyID,
			Name:        fmt.Sprintf("Leave Type %d", i),
			Code:        fmt.Sprintf("LT%d", i),
			DaysAllowed: float64(i * 5),
			IsPaid:      true,
			Status:      "active",
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	req := &dto.LeaveTypeListRequest{
		Page:     1,
		PageSize: 3,
	}

	result, err := repo.GetLeaveTypeList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetLeaveTypeList failed: %v", err)
	}

	if len(result.Data) != 3 {
		t.Errorf("expected 3 results, got %d", len(result.Data))
	}

	if result.Total < 5 {
		t.Errorf("expected total >= 5, got %d", result.Total)
	}

	if !result.HasNext {
		t.Error("expected HasNext to be true")
	}
}

func TestLeaveRepository_GetLeaveTypeList_FilterByStatus(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)

	// Create active and inactive leave types
	_, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Active Leave",
		Code:        "AL",
		DaysAllowed: 10,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Inactive Leave",
		Code:        "IL",
		DaysAllowed: 10,
		Status:      "inactive",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.LeaveTypeListRequest{
		Page:     1,
		PageSize: 10,
		Status:   "active",
	}

	result, err := repo.GetLeaveTypeList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetLeaveTypeList failed: %v", err)
	}

	for _, lt := range result.Data {
		if lt.Status != "active" {
			t.Errorf("expected status 'active', got %s", lt.Status)
		}
	}
}

func TestLeaveRepository_GetLeaveTypeList_SearchByName(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)

	_, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Searchable Leave",
		Code:        "SL",
		DaysAllowed: 10,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.LeaveTypeListRequest{
		Page:     1,
		PageSize: 10,
		Search:   "Searchable",
	}

	result, err := repo.GetLeaveTypeList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetLeaveTypeList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results for search")
	}
}

func TestLeaveRepository_UpdateLeaveType(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Original Leave",
		Code:        "OL",
		DaysAllowed: 10,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newName := "Updated Leave"
	newDays := 15.0

	updated, err := repo.UpdateLeaveType(ctx, leaveType.ID, &models.LeaveType{
		Name:        newName,
		DaysAllowed: newDays,
	})
	if err != nil {
		t.Fatalf("UpdateLeaveType failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}

	if updated.DaysAllowed != newDays {
		t.Errorf("expected days %.1f, got %.1f", newDays, updated.DaysAllowed)
	}

	// Verify persistence
	fetched, err := repo.GetLeaveTypeByID(ctx, leaveType.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if fetched.Name != newName {
		t.Errorf("persisted name should be %s, got %s", newName, fetched.Name)
	}
}

func TestLeaveRepository_DeleteLeaveType(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Delete Leave",
		Code:        "DL",
		DaysAllowed: 10,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = repo.DeleteLeaveType(ctx, leaveType.ID)
	if err != nil {
		t.Fatalf("DeleteLeaveType failed: %v", err)
	}

	_, err = repo.GetLeaveTypeByID(ctx, leaveType.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestLeaveRepository_DeleteLeaveType_NotFound(t *testing.T) {
	repo := setupLeaveRepository(t)
	ctx := context.Background()

	nonExistentID := uuid.New()
	err := repo.DeleteLeaveType(ctx, nonExistentID)
	if err == nil {
		t.Error("expected error for non-existent leave type")
	}
}

func TestLeaveRepository_CreateLeaveRequest(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	result, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("CreateLeaveRequest failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected leave request, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.EmployeeID != employeeID {
		t.Errorf("expected employee ID %v, got %v", employeeID, result.EmployeeID)
	}

	if result.Status != "pending" {
		t.Errorf("expected status 'pending', got %s", result.Status)
	}

	if result.DaysRequested != 5.0 {
		t.Errorf("expected 5 days, got %.1f", result.DaysRequested)
	}
}

func TestLeaveRepository_CreateLeaveRequest_InvalidDateFormat(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, "invalid-date", "2025-01-15", 5.0, "Vacation", nil)
	if err == nil {
		t.Error("expected error for invalid date format")
	}
}

func TestLeaveRepository_CreateLeaveRequest_EndBeforeStart(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, "2025-01-15", "2025-01-10", 5.0, "Vacation", nil)
	if err == nil {
		t.Error("expected error when end date is before start date")
	}
}

func TestLeaveRepository_CreateLeaveRequest_InsufficientBalance(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 5, // Only 5 days allowed
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 20).Format("2006-01-02")

	// Try to request more than available
	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 10.0, "Vacation", nil)
	if err == nil {
		t.Error("expected error for insufficient balance")
	}
}

func TestLeaveRepository_GetLeaveRequestByID(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	created, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result, err := repo.GetLeaveRequestByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetLeaveRequestByID failed: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("expected ID %v, got %v", created.ID, result.ID)
	}

	if result.EmployeeID != employeeID {
		t.Errorf("expected employee ID %v, got %v", employeeID, result.EmployeeID)
	}
}

func TestLeaveRepository_GetLeaveRequestList(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Create multiple requests
	for i := 1; i <= 3; i++ {
		startDate := time.Now().AddDate(0, 0, i*10).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, i*10+3).Format("2006-01-02")
		_, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 3.0, "Vacation", nil)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	req := &dto.LeaveRequestListRequest{
		Page:       1,
		PageSize:   10,
		EmployeeID: employeeID,
	}

	result, err := repo.GetLeaveRequestList(ctx, req)
	if err != nil {
		t.Fatalf("GetLeaveRequestList failed: %v", err)
	}

	if len(result.Data) != 3 {
		t.Errorf("expected 3 results, got %d", len(result.Data))
	}
}

func TestLeaveRepository_GetLeaveRequestList_FilterByStatus(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 50,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Create and approve one request
	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 13).Format("2006-01-02")
	request1, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 3.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	approverID := uuid.New()
	_, err = repo.ApproveLeaveRequest(ctx, request1.ID, approverID)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Create pending request
	startDate2 := time.Now().AddDate(0, 0, 20).Format("2006-01-02")
	endDate2 := time.Now().AddDate(0, 0, 23).Format("2006-01-02")
	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate2, endDate2, 3.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.LeaveRequestListRequest{
		Page:       1,
		PageSize:   10,
		EmployeeID: employeeID,
		Status:     "approved",
	}

	result, err := repo.GetLeaveRequestList(ctx, req)
	if err != nil {
		t.Fatalf("GetLeaveRequestList failed: %v", err)
	}

	for _, lr := range result.Data {
		if lr.Status != "approved" {
			t.Errorf("expected status 'approved', got %s", lr.Status)
		}
	}
}

func TestLeaveRepository_ApproveLeaveRequest(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	request, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	approverID := uuid.New()
	approved, err := repo.ApproveLeaveRequest(ctx, request.ID, approverID)
	if err != nil {
		t.Fatalf("ApproveLeaveRequest failed: %v", err)
	}

	if approved.Status != "approved" {
		t.Errorf("expected status 'approved', got %s", approved.Status)
	}

	if approved.ApprovedBy == nil || *approved.ApprovedBy != approverID {
		t.Error("expected approver ID to be set")
	}

	if approved.ApprovedAt == nil {
		t.Error("expected approved_at timestamp to be set")
	}

	// Verify balance updated
	balance, err := repo.GetLeaveBalance(ctx, employeeID, leaveType.ID, time.Now().Year())
	if err != nil {
		t.Fatalf("GetLeaveBalance failed: %v", err)
	}

	if balance.UsedDays != 5.0 {
		t.Errorf("expected used days 5.0, got %.1f", balance.UsedDays)
	}

	if balance.PendingDays != 0.0 {
		t.Errorf("expected pending days 0.0, got %.1f", balance.PendingDays)
	}
}

func TestLeaveRepository_RejectLeaveRequest(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	request, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rejectionReason := "Insufficient staffing"
	rejected, err := repo.RejectLeaveRequest(ctx, request.ID, rejectionReason)
	if err != nil {
		t.Fatalf("RejectLeaveRequest failed: %v", err)
	}

	if rejected.Status != "rejected" {
		t.Errorf("expected status 'rejected', got %s", rejected.Status)
	}

	if rejected.RejectionReason != rejectionReason {
		t.Errorf("expected rejection reason %s, got %s", rejectionReason, rejected.RejectionReason)
	}

	// Verify balance updated (pending days should be returned)
	balance, err := repo.GetLeaveBalance(ctx, employeeID, leaveType.ID, time.Now().Year())
	if err != nil {
		t.Fatalf("GetLeaveBalance failed: %v", err)
	}

	if balance.PendingDays != 0.0 {
		t.Errorf("expected pending days 0.0, got %.1f", balance.PendingDays)
	}
}

func TestLeaveRepository_WithdrawLeaveRequest(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	request, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	withdrawn, err := repo.WithdrawLeaveRequest(ctx, request.ID, employeeID)
	if err != nil {
		t.Fatalf("WithdrawLeaveRequest failed: %v", err)
	}

	if withdrawn.Status != "withdrawn" {
		t.Errorf("expected status 'withdrawn', got %s", withdrawn.Status)
	}

	// Verify balance updated (pending days should be returned)
	balance, err := repo.GetLeaveBalance(ctx, employeeID, leaveType.ID, time.Now().Year())
	if err != nil {
		t.Fatalf("GetLeaveBalance failed: %v", err)
	}

	if balance.PendingDays != 0.0 {
		t.Errorf("expected pending days 0.0, got %.1f", balance.PendingDays)
	}
}

func TestLeaveRepository_WithdrawLeaveRequest_WrongEmployee(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)
	otherEmployeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	request, err := repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Try to withdraw with wrong employee ID
	_, err = repo.WithdrawLeaveRequest(ctx, request.ID, otherEmployeeID)
	if err == nil {
		t.Error("expected error when withdrawing with wrong employee ID")
	}
}

func TestLeaveRepository_CheckBalance(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Create a leave request to initialize balance
	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	available, err := repo.CheckBalance(ctx, employeeID, leaveType.ID, time.Now().Year())
	if err != nil {
		t.Fatalf("CheckBalance failed: %v", err)
	}

	// Should be 20 total - 5 pending = 15 available
	if available != 15.0 {
		t.Errorf("expected available days 15.0, got %.1f", available)
	}
}

func TestLeaveRepository_CheckBalance_NoBalance(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)
	leaveTypeID := uuid.New()

	available, err := repo.CheckBalance(ctx, employeeID, leaveTypeID, time.Now().Year())
	if err != nil {
		t.Fatalf("CheckBalance failed: %v", err)
	}

	// Should return 0 when no balance exists
	if available != 0.0 {
		t.Errorf("expected available days 0.0, got %.1f", available)
	}
}

func TestLeaveRepository_GetEmployeeBalances(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	// Create multiple leave types
	leaveType1, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	leaveType2, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Sick Leave",
		Code:        "SL",
		DaysAllowed: 10,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Create requests to initialize balances
	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 12).Format("2006-01-02")

	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType1.ID, startDate, endDate, 2.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType2.ID, startDate, endDate, 2.0, "Sick", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	balances, err := repo.GetEmployeeBalances(ctx, employeeID, time.Now().Year())
	if err != nil {
		t.Fatalf("GetEmployeeBalances failed: %v", err)
	}

	if len(balances) != 2 {
		t.Errorf("expected 2 balances, got %d", len(balances))
	}

	for _, balance := range balances {
		if balance.EmployeeID != employeeID {
			t.Errorf("expected employee ID %v, got %v", employeeID, balance.EmployeeID)
		}
		if balance.PendingDays != 2.0 {
			t.Errorf("expected pending days 2.0, got %.1f", balance.PendingDays)
		}
	}
}

func TestLeaveRepository_GetLeaveBalance(t *testing.T) {
	repo := setupLeaveRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := createTestCompany(t, pool)
	defer cleanupLeaveTestData(ctx, pool, companyID)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	employeeID := createTestEmployee(t, pool, companyID)

	leaveType, err := repo.CreateLeaveType(ctx, &models.LeaveType{
		CompanyID:   companyID,
		Name:        "Annual Leave",
		Code:        "AL",
		DaysAllowed: 20,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Create request to initialize balance
	startDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")

	_, err = repo.CreateLeaveRequest(ctx, employeeID, leaveType.ID, startDate, endDate, 5.0, "Vacation", nil)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	balance, err := repo.GetLeaveBalance(ctx, employeeID, leaveType.ID, time.Now().Year())
	if err != nil {
		t.Fatalf("GetLeaveBalance failed: %v", err)
	}

	if balance.TotalDays != 20.0 {
		t.Errorf("expected total days 20.0, got %.1f", balance.TotalDays)
	}

	if balance.PendingDays != 5.0 {
		t.Errorf("expected pending days 5.0, got %.1f", balance.PendingDays)
	}

	if balance.UsedDays != 0.0 {
		t.Errorf("expected used days 0.0, got %.1f", balance.UsedDays)
	}
}