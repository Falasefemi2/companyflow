package services

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
)

func TestEmployeeService_CreateEmployee(t *testing.T) {
	service := setupEmployeeService(t)
	ctx := context.Background()

	companyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleID := uuid.MustParse("b2711d17-5b6d-4e9a-98c6-bc654184cd4f")
	departmentID := uuid.New().String()
	designationID := uuid.New().String()
	levelID := uuid.New().String()

	pool := setupTestDB(t)
	if err := cleanupEmployeeTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	req := &dto.CreateEmployeeRequest{
		Email:                 "service.test@example.com",
		Password:              "SecurePassword123!",
		Phone:                 "+1234567890",
		FirstName:             "Service",
		LastName:              "Test",
		DateOfBirth:           "1990-01-15",
		EmployeeCode:          "EMP001",
		DepartmentID:          departmentID,
		DesignationID:         designationID,
		LevelID:               levelID,
		RoleID:                roleID.String(),
		Status:                "active",
		EmploymentType:        "full_time",
		HireDate:              "2024-01-15",
		Gender:                "Male",
		Address:               "123 Test Street",
		EmergencyContactName:  "John Doe",
		EmergencyContactPhone: "+0987654321",
		ProfileImageUrl:       "https://example.com/image.jpg",
	}

	result, err := service.CreateEmployee(ctx, req)
	if err != nil {
		t.Fatalf("CreateEmployee failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected employee response, got nil")
	}

	if result.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, result.Email)
	}

	if result.FirstName != req.FirstName {
		t.Errorf("expected first_name %s, got %s", req.FirstName, result.FirstName)
	}

	if result.EmployeeCode != req.EmployeeCode {
		t.Errorf("expected employee_code %s, got %s", req.EmployeeCode, result.EmployeeCode)
	}

	if result.CompanyID != companyID.String() {
		t.Errorf("expected company_id %s, got %s", companyID.String(), result.CompanyID)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps should not be zero")
	}
}
