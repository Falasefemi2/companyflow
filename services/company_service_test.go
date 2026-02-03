package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/utils"
)

func TestCompanyService_CreateCompany(t *testing.T) {
	service := setupCompanyService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	if err := cleanupCompanyTestData(ctx, pool, "svc-company-%"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	req := &dto.CreateCompanyRequest{
		Name:               "Service Company",
		Slug:               fmt.Sprintf("svc-company-%d", time.Now().UnixNano()),
		Industry:           "Technology",
		Country:            "Nigeria",
		Timezone:           "Africa/Lagos",
		Currency:           "NGN",
		RegistrationNumber: "RC-100001",
		TaxID:              "TIN-100001",
		Address:            "10 Service Street",
		Phone:              "+2348000000000",
		LogoURL:            "https://example.com/logo.png",
		Status:             "active",
	}

	result, err := service.CreateCompany(ctx, req)
	if err != nil {
		t.Fatalf("CreateCompany failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected company response, got nil")
	}

	if result.Slug != req.Slug {
		t.Errorf("expected slug %s, got %s", req.Slug, result.Slug)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps should not be zero")
	}
}

func TestCompanyService_GetCompanyByID(t *testing.T) {
	service := setupCompanyService(t)
	ctx := context.Background()

	req := &dto.CreateCompanyRequest{
		Name: "Get Company",
		Slug: fmt.Sprintf("svc-get-company-%d", time.Now().UnixNano()),
	}

	created, err := service.CreateCompany(ctx, req)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	companyID := uuid.MustParse(created.ID)

	result, err := service.GetCompanyByID(ctx, companyID)
	if err != nil {
		t.Fatalf("GetCompanyByID failed: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("expected %s, got %s", created.ID, result.ID)
	}

	if result.Slug != created.Slug {
		t.Errorf("expected slug %s, got %s", created.Slug, result.Slug)
	}
}

func TestCompanyService_GetCompanyList(t *testing.T) {
	service := setupCompanyService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	if err := cleanupCompanyTestData(ctx, pool, "svc-list-company-%"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	for i := 1; i <= 5; i++ {
		_, err := service.CreateCompany(ctx, &dto.CreateCompanyRequest{
			Name: "List Company",
			Slug: fmt.Sprintf("svc-list-company-%d-%d", i, time.Now().UnixNano()),
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

	result, err := service.GetCompanyList(ctx, req)
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

func TestCompanyService_UpdateCompany(t *testing.T) {
	service := setupCompanyService(t)
	ctx := context.Background()

	created, err := service.CreateCompany(ctx, &dto.CreateCompanyRequest{
		Name: "Original Company",
		Slug: fmt.Sprintf("svc-update-company-%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newName := "Updated Company"
	newStatus := "inactive"

	updated, err := service.UpdateCompany(ctx, uuid.MustParse(created.ID), &dto.UpdateCompanyRequest{
		Name:   &newName,
		Status: &newStatus,
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
}

func TestCompanyService_DeleteCompany(t *testing.T) {
	service := setupCompanyService(t)
	ctx := context.Background()

	created, err := service.CreateCompany(ctx, &dto.CreateCompanyRequest{
		Name: "Delete Company",
		Slug: fmt.Sprintf("svc-delete-company-%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	companyID := uuid.MustParse(created.ID)

	err = service.DeleteCompany(ctx, companyID, true)
	if err != nil {
		t.Fatalf("soft delete failed: %v", err)
	}

	err = service.DeleteCompany(ctx, companyID, false)
	if err != nil {
		t.Fatalf("hard delete failed: %v", err)
	}
}
