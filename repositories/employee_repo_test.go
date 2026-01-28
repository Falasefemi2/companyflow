package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/utils"
)

func TestEmployeeRepository_CreateEmployee(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleID := uuid.MustParse("b2711d17-5b6d-4e9a-98c6-bc654184cd4f")

	employee := &models.Employee{
		CompanyID:             companyID,
		Email:                 "test.employee@example.com",
		PasswordHash:          "hashed_password",
		Phone:                 "+1234567890",
		FirstName:             "Test",
		LastName:              "Employee",
		EmployeeCode:          "TEST001",
		RoleID:                roleID,
		Status:                "active",
		EmploymentType:        "full_time",
		HireDate:              time.Now(),
		Gender:                "Male",
		Address:               "123 Test Street",
		EmergencyContactName:  "Emergency",
		EmergencyContactPhone: "+0987654321",
	}

	result, err := repo.CreateEmployee(ctx, employee)
	if err != nil {
		t.Fatalf("CreateEmployee failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected employee, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.Email != employee.Email {
		t.Errorf("expected email %s, got %s", employee.Email, result.Email)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestEmployeeRepository_GetEmployeeByID(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleID := uuid.MustParse("b2711d17-5b6d-4e9a-98c6-bc654184cd4f")

	employee := &models.Employee{
		CompanyID:      companyID,
		Email:          "retrieve@example.com",
		PasswordHash:   "hashed",
		Phone:          "+1234567890",
		FirstName:      "Retrieve",
		LastName:       "Test",
		EmployeeCode:   "RETRIEVE01",
		RoleID:         roleID,
		Status:         "active",
		EmploymentType: "full_time",
		HireDate:       time.Now(),
	}

	created, err := repo.CreateEmployee(ctx, employee)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result, err := repo.GetEmployeeByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetEmployeeByID failed: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("expected %v, got %v", created.ID, result.ID)
	}
}

func TestEmployeeRepository_GetEmployeeList(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleID := uuid.MustParse("b2711d17-5b6d-4e9a-98c6-bc654184cd4f")

	for i := 1; i <= 5; i++ {
		_, err := repo.CreateEmployee(ctx, &models.Employee{
			CompanyID:      companyID,
			Email:          fmt.Sprintf("list%d@example.com", i),
			PasswordHash:   "hashed",
			Phone:          "+1234567890",
			FirstName:      "Employee",
			LastName:       fmt.Sprintf("Test%d", i),
			EmployeeCode:   fmt.Sprintf("LIST%03d", i),
			RoleID:         roleID,
			Status:         "active",
			EmploymentType: "full_time",
			HireDate:       time.Now(),
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	req := &dto.EmployeeListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 2,
		},
	}

	result, err := repo.GetEmployeeList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetEmployeeList failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("expected 2, got %d", len(result.Data))
	}

	if result.Total < 5 {
		t.Errorf("expected total >= 5, got %d", result.Total)
	}

	if !result.HasNext {
		t.Error("expected HasNext true")
	}
}

func TestEmployeeRepository_GetEmployeeList_FilterStatus(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleID := uuid.MustParse("b2711d17-5b6d-4e9a-98c6-bc654184cd4f")

	_, err := repo.CreateEmployee(ctx, &models.Employee{
		CompanyID:      companyID,
		Email:          "active@example.com",
		PasswordHash:   "hashed",
		Phone:          "+1234567890",
		FirstName:      "Active",
		LastName:       "User",
		EmployeeCode:   "ACTIVE01",
		RoleID:         roleID,
		Status:         "active",
		EmploymentType: "full_time",
		HireDate:       time.Now(),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.EmployeeListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Status: "active",
	}

	result, err := repo.GetEmployeeList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetEmployeeList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results")
	}

	for _, e := range result.Data {
		if e.Status != "active" {
			t.Errorf("unexpected status %s", e.Status)
		}
	}
}

func TestEmployeeRepository_DeleteEmployee(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleID := uuid.MustParse("b2711d17-5b6d-4e9a-98c6-bc654184cd4f")

	employee, err := repo.CreateEmployee(ctx, &models.Employee{
		CompanyID:      companyID,
		Email:          "delete@example.com",
		PasswordHash:   "hashed",
		Phone:          "+1234567890",
		FirstName:      "Delete",
		LastName:       "Test",
		EmployeeCode:   "DEL01",
		RoleID:         roleID,
		Status:         "active",
		EmploymentType: "full_time",
		HireDate:       time.Now(),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = repo.DeleteEmployee(ctx, employee.ID.String(), false)
	if err != nil {
		t.Fatalf("soft delete failed: %v", err)
	}

	updated, err := repo.GetEmployeeByID(ctx, employee.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if updated.Status != "inactive" {
		t.Errorf("expected inactive, got %s", updated.Status)
	}

	err = repo.DeleteEmployee(ctx, employee.ID.String(), true)
	if err != nil {
		t.Fatalf("hard delete failed: %v", err)
	}

	_, err = repo.GetEmployeeByID(ctx, employee.ID)
	if err == nil {
		t.Error("expected error after hard delete")
	}
}
