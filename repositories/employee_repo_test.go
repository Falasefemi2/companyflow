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

const (
	testCompanyID = "550e8400-e29b-41d4-a716-446655440000"
	testRoleID    = "b2711d17-5b6d-4e9a-98c6-bc654184cd4f"
)

func cleanupEmployeeDependencies(t *testing.T, ctx context.Context, poolID string) {
	t.Helper()
	pool := setupTestDB(t)

	if err := cleanupEmployeeTestData(ctx, pool, poolID); err != nil {
		t.Fatalf("cleanup employees failed: %v", err)
	}
	if err := cleanupDesignationTestData(ctx, pool, poolID); err != nil {
		t.Fatalf("cleanup designations failed: %v", err)
	}
	if err := cleanupDepartmentTestData(ctx, pool, poolID); err != nil {
		t.Fatalf("cleanup departments failed: %v", err)
	}
	if err := cleanupLevelTestData(ctx, pool, poolID); err != nil {
		t.Fatalf("cleanup levels failed: %v", err)
	}
}

func setupEmployeeDependencies(t *testing.T, ctx context.Context, companyID uuid.UUID) (*uuid.UUID, *uuid.UUID, *uuid.UUID) {
	t.Helper()

	departmentRepo := setupDepartmentRepository(t)
	levelRepo := setupLevelRepository(t)
	designationRepo := setupDesignationRepository(t)

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Level-%d", time.Now().UnixNano()),
		HierarchyLevel: 1,
		MinSalary:      ptrFloat(2000000.00),
		MaxSalary:      ptrFloat(8000000.00),
		Description:    "Employee test level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	department, err := departmentRepo.CreateDepartment(ctx, &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Department-%d", time.Now().UnixNano()),
		Code:        "EMP",
		Description: "Employee test department",
		CostCenter:  "CC-EMP",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed (create department): %v", err)
	}

	designation, err := designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID:    companyID,
		Name:         fmt.Sprintf("Designation-%d", time.Now().UnixNano()),
		Description:  "Employee test designation",
		LevelID:      &level.ID,
		DepartmentID: &department.ID,
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("setup failed (create designation): %v", err)
	}

	return &department.ID, &designation.ID, &level.ID
}

func TestEmployeeRepository_CreateEmployee(t *testing.T) {
	repo := setupEmployeeRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)
	roleID := uuid.MustParse(testRoleID)

	cleanupEmployeeDependencies(t, ctx, companyID.String())

	deptID, designationID, levelID := setupEmployeeDependencies(t, ctx, companyID)

	uniqueEmail := fmt.Sprintf("test.employee.%d@example.com", time.Now().UnixNano())
	employee := &models.Employee{
		CompanyID:             companyID,
		Email:                 uniqueEmail,
		PasswordHash:          "hashed_password",
		Phone:                 "+1234567890",
		FirstName:             "Test",
		LastName:              "Employee",
		EmployeeCode:          "TEST001",
		DepartmentID:          deptID,
		DesignationID:         designationID,
		LevelID:               levelID,
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

	if result.DepartmentID == nil || *result.DepartmentID != *deptID {
		t.Error("expected DepartmentID to be set")
	}

	if result.DesignationID == nil || *result.DesignationID != *designationID {
		t.Error("expected DesignationID to be set")
	}

	if result.LevelID == nil || *result.LevelID != *levelID {
		t.Error("expected LevelID to be set")
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestEmployeeRepository_GetEmployeeByID(t *testing.T) {
	repo := setupEmployeeRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)
	roleID := uuid.MustParse(testRoleID)

	cleanupEmployeeDependencies(t, ctx, companyID.String())

	deptID, designationID, levelID := setupEmployeeDependencies(t, ctx, companyID)

	employee := &models.Employee{
		CompanyID:      companyID,
		Email:          fmt.Sprintf("retrieve.%d@example.com", time.Now().UnixNano()),
		PasswordHash:   "hashed",
		Phone:          "+1234567890",
		FirstName:      "Retrieve",
		LastName:       "Test",
		EmployeeCode:   fmt.Sprintf("RETRIEVE%d", time.Now().Unix()),
		DepartmentID:   deptID,
		DesignationID:  designationID,
		LevelID:        levelID,
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

	if result.Email != created.Email {
		t.Errorf("expected email %s, got %s", created.Email, result.Email)
	}

	if result.DepartmentID == nil || *result.DepartmentID != *deptID {
		t.Error("expected DepartmentID to match")
	}

	if result.DesignationID == nil || *result.DesignationID != *designationID {
		t.Error("expected DesignationID to match")
	}

	if result.LevelID == nil || *result.LevelID != *levelID {
		t.Error("expected LevelID to match")
	}
}

func TestEmployeeRepository_GetEmployeeList(t *testing.T) {
	repo := setupEmployeeRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)
	roleID := uuid.MustParse(testRoleID)

	cleanupEmployeeDependencies(t, ctx, companyID.String())

	deptID, designationID, levelID := setupEmployeeDependencies(t, ctx, companyID)

	for i := 1; i <= 5; i++ {
		_, err := repo.CreateEmployee(ctx, &models.Employee{
			CompanyID:      companyID,
			Email:          fmt.Sprintf("list%d.%d@example.com", i, time.Now().UnixNano()),
			PasswordHash:   "hashed",
			Phone:          "+1234567890",
			FirstName:      "Employee",
			LastName:       fmt.Sprintf("Test%d", i),
			EmployeeCode:   fmt.Sprintf("LIST%03d", i),
			DepartmentID:   deptID,
			DesignationID:  designationID,
			LevelID:        levelID,
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
	repo := setupEmployeeRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)
	roleID := uuid.MustParse(testRoleID)

	cleanupEmployeeDependencies(t, ctx, companyID.String())

	deptID, designationID, levelID := setupEmployeeDependencies(t, ctx, companyID)

	_, err := repo.CreateEmployee(ctx, &models.Employee{
		CompanyID:      companyID,
		Email:          fmt.Sprintf("active.%d@example.com", time.Now().UnixNano()),
		PasswordHash:   "hashed",
		Phone:          "+1234567890",
		FirstName:      "Active",
		LastName:       "User",
		EmployeeCode:   fmt.Sprintf("ACTIVE%d", time.Now().Unix()),
		DepartmentID:   deptID,
		DesignationID:  designationID,
		LevelID:        levelID,
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
	repo := setupEmployeeRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)
	roleID := uuid.MustParse(testRoleID)

	cleanupEmployeeDependencies(t, ctx, companyID.String())

	deptID, designationID, levelID := setupEmployeeDependencies(t, ctx, companyID)

	employee, err := repo.CreateEmployee(ctx, &models.Employee{
		CompanyID:      companyID,
		Email:          fmt.Sprintf("delete.%d@example.com", time.Now().UnixNano()),
		PasswordHash:   "hashed",
		Phone:          "+1234567890",
		FirstName:      "Delete",
		LastName:       "Test",
		EmployeeCode:   fmt.Sprintf("DEL%d", time.Now().Unix()),
		DepartmentID:   deptID,
		DesignationID:  designationID,
		LevelID:        levelID,
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
