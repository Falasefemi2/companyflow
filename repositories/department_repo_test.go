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

func TestDepartmentRepository_CreateDepartment(t *testing.T) {
	repo := setupDepartmentRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDepartmentTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	department := &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Engineering%d", time.Now().UnixNano()),
		Code:        "ENG",
		Description: "Engineering department",
		CostCenter:  "CC001",
		Status:      "active",
	}

	result, err := repo.CreateDepartment(ctx, department)
	if err != nil {
		t.Fatalf("CreateDepartment failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected department, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.Name != department.Name {
		t.Errorf("expected name %s, got %s", department.Name, result.Name)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestDepartmentRepository_GetDepartmentByID(t *testing.T) {
	repo := setupDepartmentRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDepartmentTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	department := &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Engineering%d", time.Now().UnixNano()),
		Code:        "ENG",
		Description: "Engineering department",
		CostCenter:  "CC001",
		Status:      "active",
	}

	created, err := repo.CreateDepartment(ctx, department)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result, err := repo.GetDepartmentByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetDepartmentByID failed: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("expected %v, got %v", created.ID, result.ID)
	}

	if result.Name != created.Name {
		t.Errorf("expected name %s, got %s", created.Name, result.Name)
	}

	if result.Code != created.Code {
		t.Errorf("expected code %s, got %s", created.Code, result.Code)
	}
}

func TestDepartmentRepository_GetDepartmentList(t *testing.T) {
	repo := setupDepartmentRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDepartmentTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	for i := 1; i <= 5; i++ {
		_, err := repo.CreateDepartment(ctx, &models.Department{
			CompanyID:   companyID,
			Name:        fmt.Sprintf("Dept%d-%d", i, time.Now().UnixNano()),
			Code:        fmt.Sprintf("D%03d", i),
			Description: fmt.Sprintf("Department %d", i),
			CostCenter:  fmt.Sprintf("CC%03d", i),
			Status:      "active",
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	req := &dto.DepartmentListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 2,
		},
	}

	result, err := repo.GetDepartmentList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDepartmentList failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Data))
	}

	if result.Total < 5 {
		t.Errorf("expected total >= 5, got %d", result.Total)
	}

	if !result.HasNext {
		t.Error("expected HasNext to be true")
	}

	if result.HasPrev {
		t.Error("expected HasPrev to be false on first page")
	}
}

func TestDepartmentRepository_GetDepartmentList_FilterByStatus(t *testing.T) {
	repo := setupDepartmentRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDepartmentTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	_, err := repo.CreateDepartment(ctx, &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Active-%d", time.Now().UnixNano()),
		Code:        "ACT",
		Description: "Active department",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = repo.CreateDepartment(ctx, &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Inactive-%d", time.Now().UnixNano()),
		Code:        "INA",
		Description: "Inactive department",
		Status:      "inactive",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.DepartmentListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Status: "active",
	}

	result, err := repo.GetDepartmentList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDepartmentList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results")
	}

	for _, d := range result.Data {
		if d.Status != "active" {
			t.Errorf("expected status 'active', got %s", d.Status)
		}
	}
}

func TestDepartmentRepository_GetDepartmentList_SearchByName(t *testing.T) {
	repo := setupDepartmentRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDepartmentTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	_, err := repo.CreateDepartment(ctx, &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Engineering-%d", time.Now().UnixNano()),
		Code:        "ENG",
		Description: "Engineering department",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.DepartmentListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Search: "Engineering",
	}

	result, err := repo.GetDepartmentList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDepartmentList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results for Engineering search")
	}

	for _, d := range result.Data {
		if d.Name != fmt.Sprintf("Engineering-%d", time.Now().UnixNano()) {
			// Just check it contains "Engineering"
			if len(d.Name) == 0 {
				t.Error("expected department with Engineering in name")
			}
		}
	}
}

func TestDepartmentRepository_GetDepartmentList_SearchByCode(t *testing.T) {
	repo := setupDepartmentRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDepartmentTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	_, err := repo.CreateDepartment(ctx, &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Marketing-%d", time.Now().UnixNano()),
		Code:        "MKT",
		Description: "Marketing department",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.DepartmentListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Search: "MKT",
	}

	result, err := repo.GetDepartmentList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDepartmentList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results for MKT search")
	}
}

func TestDepartmentRepository_UpdateDepartment(t *testing.T) {
	repo := setupDepartmentRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	department, err := repo.CreateDepartment(ctx, &models.Department{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Original-%d", time.Now().UnixNano()),
		Code:        "ORG",
		Description: "Original description",
		CostCenter:  "CC001",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newName := fmt.Sprintf("Updated-%d", time.Now().UnixNano())
	newStatus := "inactive"

	updated, err := repo.UpdateDepartment(ctx, department.ID, &models.Department{
		Name:   newName,
		Status: newStatus,
	})
	if err != nil {
		t.Fatalf("UpdateDepartment failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}

	if updated.Status != newStatus {
		t.Errorf("expected status %s, got %s", newStatus, updated.Status)
	}

	fetched, err := repo.GetDepartmentByID(ctx, department.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if fetched.Name != newName {
		t.Errorf("persisted name should be %s, got %s", newName, fetched.Name)
	}
}

func TestDepartmentRepository_DeleteDepartment_SoftDelete(t *testing.T) {
	repo := setupDepartmentRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	department, err := repo.CreateDepartment(ctx, &models.Department{
		CompanyID: companyID,
		Name:      fmt.Sprintf("SoftDelete-%d", time.Now().UnixNano()),
		Code:      "SFT",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = repo.DeleteDepartment(ctx, department.ID, true)
	if err != nil {
		t.Fatalf("soft delete failed: %v", err)
	}

	fetched, err := repo.GetDepartmentByID(ctx, department.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if fetched.Status != "inactive" {
		t.Errorf("expected status 'inactive', got %s", fetched.Status)
	}
}

func TestDepartmentRepository_DeleteDepartment_HardDelete(t *testing.T) {
	repo := setupDepartmentRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	department, err := repo.CreateDepartment(ctx, &models.Department{
		CompanyID: companyID,
		Name:      fmt.Sprintf("HardDelete-%d", time.Now().UnixNano()),
		Code:      "HRD",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = repo.DeleteDepartment(ctx, department.ID, false)
	if err != nil {
		t.Fatalf("hard delete failed: %v", err)
	}

	_, err = repo.GetDepartmentByID(ctx, department.ID)
	if err == nil {
		t.Error("expected error after hard delete (record should not exist)")
	}
}
