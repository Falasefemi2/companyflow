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

func TestCompanyRepository_CreateCompany(t *testing.T) {
	repo := setupCompanyRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	if err := cleanupCompanyTestData(ctx, pool, "test-company-%"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	company := &models.Company{
		Name:               "Test Company",
		Slug:               fmt.Sprintf("test-company-%d", time.Now().UnixNano()),
		Industry:           "Technology",
		Country:            "Nigeria",
		Timezone:           "Africa/Lagos",
		Currency:           "NGN",
		RegistrationNumber: "RC-123456",
		TaxID:              "TIN-987654",
		Address:            "123 Test Street",
		Phone:              "+2348000000000",
		LogoURL:            "https://example.com/logo.png",
		Status:             "active",
		Settings:           []byte(`{}`),
	}

	result, err := repo.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("CreateCompany failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected company, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.Slug != company.Slug {
		t.Errorf("expected slug %s, got %s", company.Slug, result.Slug)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestCompanyRepository_GetCompanyByID(t *testing.T) {
	repo := setupCompanyRepository(t)
	ctx := context.Background()

	company := &models.Company{
		Name:     "Retrieve Company",
		Slug:     fmt.Sprintf("retrieve-company-%d", time.Now().UnixNano()),
		Status:   "active",
		Settings: []byte(`{}`),
	}

	created, err := repo.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result, err := repo.GetCompanyByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetCompanyByID failed: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("expected %v, got %v", created.ID, result.ID)
	}

	if result.Slug != created.Slug {
		t.Errorf("expected slug %s, got %s", created.Slug, result.Slug)
	}
}

func TestCompanyRepository_GetCompanyList(t *testing.T) {
	repo := setupCompanyRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	if err := cleanupCompanyTestData(ctx, pool, "list-company-%"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	for i := 1; i <= 5; i++ {
		_, err := repo.CreateCompany(ctx, &models.Company{
			Name:     fmt.Sprintf("List Company %d", i),
			Slug:     fmt.Sprintf("list-company-%d-%d", i, time.Now().UnixNano()),
			Status:   "active",
			Settings: []byte(`{}`),
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	req := &dto.CompanyListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 2,
		},
	}

	result, err := repo.GetCompanyList(ctx, req)
	if err != nil {
		t.Fatalf("GetCompanyList failed: %v", err)
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
}

func TestCompanyRepository_GetCompanyList_FilterStatus(t *testing.T) {
	repo := setupCompanyRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	if err := cleanupCompanyTestData(ctx, pool, "status-company-%"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	_, err := repo.CreateCompany(ctx, &models.Company{
		Name:     "Active Company",
		Slug:     fmt.Sprintf("status-company-active-%d", time.Now().UnixNano()),
		Status:   "active",
		Settings: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = repo.CreateCompany(ctx, &models.Company{
		Name:     "Inactive Company",
		Slug:     fmt.Sprintf("status-company-inactive-%d", time.Now().UnixNano()),
		Status:   "inactive",
		Settings: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.CompanyListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Status: "active",
	}

	result, err := repo.GetCompanyList(ctx, req)
	if err != nil {
		t.Fatalf("GetCompanyList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results")
	}

	for _, c := range result.Data {
		if c.Status != "active" {
			t.Errorf("expected status 'active', got %s", c.Status)
		}
	}
}

func TestCompanyRepository_GetCompanyList_SearchByName(t *testing.T) {
	repo := setupCompanyRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	if err := cleanupCompanyTestData(ctx, pool, "search-company-%"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	_, err := repo.CreateCompany(ctx, &models.Company{
		Name:     "Searchable Company",
		Slug:     fmt.Sprintf("search-company-%d", time.Now().UnixNano()),
		Status:   "active",
		Settings: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.CompanyListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Search: "Searchable",
	}

	result, err := repo.GetCompanyList(ctx, req)
	if err != nil {
		t.Fatalf("GetCompanyList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results for Searchable search")
	}
}

func TestCompanyRepository_UpdateCompany(t *testing.T) {
	repo := setupCompanyRepository(t)
	ctx := context.Background()

	company, err := repo.CreateCompany(ctx, &models.Company{
		Name:     "Original Company",
		Slug:     fmt.Sprintf("update-company-%d", time.Now().UnixNano()),
		Status:   "active",
		Settings: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newName := "Updated Company"
	newStatus := "inactive"

	updated, err := repo.UpdateCompany(ctx, company.ID, &models.Company{
		Name:   newName,
		Status: newStatus,
	})
	if err != nil {
		t.Fatalf("UpdateCompany failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}

	if updated.Status != newStatus {
		t.Errorf("expected status %s, got %s", newStatus, updated.Status)
	}

	fetched, err := repo.GetCompanyByID(ctx, company.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if fetched.Name != newName {
		t.Errorf("persisted name should be %s, got %s", newName, fetched.Name)
	}
}

func TestCompanyRepository_DeleteCompany(t *testing.T) {
	repo := setupCompanyRepository(t)
	ctx := context.Background()

	company, err := repo.CreateCompany(ctx, &models.Company{
		Name:     "Delete Company",
		Slug:     fmt.Sprintf("delete-company-%d", time.Now().UnixNano()),
		Status:   "active",
		Settings: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = repo.DeleteCompany(ctx, company.ID, true)
	if err != nil {
		t.Fatalf("soft delete failed: %v", err)
	}

	updated, err := repo.GetCompanyByID(ctx, company.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if updated.Status != "inactive" {
		t.Errorf("expected inactive, got %s", updated.Status)
	}

	err = repo.DeleteCompany(ctx, company.ID, false)
	if err != nil {
		t.Fatalf("hard delete failed: %v", err)
	}

	_, err = repo.GetCompanyByID(ctx, company.ID)
	if err == nil {
		t.Error("expected error after hard delete")
	}
}
