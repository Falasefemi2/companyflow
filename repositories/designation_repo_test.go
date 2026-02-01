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

func TestDesignationRepository_CreateDesignation(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDesignationTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Senior-%d", time.Now().UnixNano()),
		HierarchyLevel: 3,
		MinSalary:      ptrFloat(5000000.00),
		MaxSalary:      ptrFloat(15000000.00),
		Description:    "Senior level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	designation := &models.Designation{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Senior Backend Developer-%d", time.Now().UnixNano()),
		Description: "Senior backend development role",
		LevelID:     &level.ID,
		Status:      "active",
	}

	result, err := designationRepo.CreateDesignation(ctx, designation)
	if err != nil {
		t.Fatalf("CreateDesignation failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected designation, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.Name != designation.Name {
		t.Errorf("expected name %s, got %s", designation.Name, result.Name)
	}

	if result.LevelID == nil || *result.LevelID != level.ID {
		t.Error("expected LevelID to match created level")
	}

	if result.Status != "active" {
		t.Errorf("expected status 'active', got %s", result.Status)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestDesignationRepository_GetDesignationByID(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDesignationTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Junior-%d", time.Now().UnixNano()),
		HierarchyLevel: 1,
		Description:    "Junior level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	designation, err := designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Junior Developer-%d", time.Now().UnixNano()),
		Description: "Junior development role",
		LevelID:     &level.ID,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed (create designation): %v", err)
	}

	result, err := designationRepo.GetDesignationByID(ctx, designation.ID)
	// Assert
	if err != nil {
		t.Fatalf("GetDesignationByID failed: %v", err)
	}

	if result.ID != designation.ID {
		t.Errorf("expected %v, got %v", designation.ID, result.ID)
	}

	if result.Name != designation.Name {
		t.Errorf("expected name %s, got %s", designation.Name, result.Name)
	}

	if result.LevelID == nil || *result.LevelID != level.ID {
		t.Error("expected LevelID to match")
	}

	if result.Status != "active" {
		t.Errorf("expected status 'active', got %s", result.Status)
	}
}

func TestDesignationRepository_GetDesignationList(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDesignationTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Mid-Level-%d", time.Now().UnixNano()),
		HierarchyLevel: 2,
		Description:    "Mid-level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	for i := 1; i <= 5; i++ {
		_, err := designationRepo.CreateDesignation(ctx, &models.Designation{
			CompanyID:   companyID,
			Name:        fmt.Sprintf("Designation%d-%d", i, time.Now().UnixNano()),
			Description: fmt.Sprintf("Designation %d", i),
			LevelID:     &level.ID,
			Status:      "active",
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	req := &dto.DesignationListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 2,
		},
	}

	result, err := designationRepo.GetDesignationList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDesignationList failed: %v", err)
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

func TestDesignationRepository_GetDesignationList_FilterStatus(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDesignationTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Manager-%d", time.Now().UnixNano()),
		HierarchyLevel: 4,
		Description:    "Manager level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	_, err = designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Active Manager-%d", time.Now().UnixNano()),
		Description: "Active manager role",
		LevelID:     &level.ID,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Inactive Manager-%d", time.Now().UnixNano()),
		Description: "Inactive manager role",
		LevelID:     &level.ID,
		Status:      "inactive",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.DesignationListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Status: "active",
	}

	result, err := designationRepo.GetDesignationList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDesignationList failed: %v", err)
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

func TestDesignationRepository_GetDesignationList_SearchByName(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDesignationTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Executive-%d", time.Now().UnixNano()),
		HierarchyLevel: 5,
		Description:    "Executive level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	_, err = designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Chief Technology Officer-%d", time.Now().UnixNano()),
		Description: "CTO role",
		LevelID:     &level.ID,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.DesignationListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Search: "Chief",
	}

	result, err := designationRepo.GetDesignationList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDesignationList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results for 'Chief' search")
	}
}

func TestDesignationRepository_GetDesignationList_FilterByLevel(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupDesignationTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level1, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Level1-%d", time.Now().UnixNano()),
		HierarchyLevel: 1,
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	level2, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Level2-%d", time.Now().UnixNano()),
		HierarchyLevel: 2,
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID: companyID,
		Name:      fmt.Sprintf("Designations for Level1-%d", time.Now().UnixNano()),
		LevelID:   &level1.ID,
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID: companyID,
		Name:      fmt.Sprintf("Designations for Level2-%d", time.Now().UnixNano()),
		LevelID:   &level2.ID,
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.DesignationListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		LevelID: level1.ID.String(),
	}

	result, err := designationRepo.GetDesignationList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetDesignationList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results for level filter")
	}

	for _, d := range result.Data {
		if d.LevelID == nil || *d.LevelID != level1.ID {
			t.Errorf("expected LevelID to match filter")
		}
	}
}

func TestDesignationRepository_UpdateDesignation(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Update-Level-%d", time.Now().UnixNano()),
		HierarchyLevel: 2,
		Description:    "Update test level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	designation, err := designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Original-Designation-%d", time.Now().UnixNano()),
		Description: "Original description",
		LevelID:     &level.ID,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed (create designation): %v", err)
	}

	newName := fmt.Sprintf("Updated-Designation-%d", time.Now().UnixNano())
	newStatus := "inactive"

	updated, err := designationRepo.UpdateDesignation(ctx, designation.ID, &models.Designation{
		Name:   newName,
		Status: newStatus,
	})
	if err != nil {
		t.Fatalf("UpdateDesignation failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}

	if updated.Status != newStatus {
		t.Errorf("expected status %s, got %s", newStatus, updated.Status)
	}

	fetched, err := designationRepo.GetDesignationByID(ctx, designation.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if fetched.Name != newName {
		t.Errorf("persisted name should be %s, got %s", newName, fetched.Name)
	}
}

func TestDesignationRepository_DeleteDesignation(t *testing.T) {
	designationRepo := setupDesignationRepository(t)
	levelRepo := setupLevelRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	level, err := levelRepo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Delete-Level-%d", time.Now().UnixNano()),
		HierarchyLevel: 1,
		Description:    "Delete test level",
	})
	if err != nil {
		t.Fatalf("setup failed (create level): %v", err)
	}

	designation, err := designationRepo.CreateDesignation(ctx, &models.Designation{
		CompanyID:   companyID,
		Name:        fmt.Sprintf("Delete-Designation-%d", time.Now().UnixNano()),
		Description: "To be deleted",
		LevelID:     &level.ID,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("setup failed (create designation): %v", err)
	}

	err = designationRepo.DeleteDesignation(ctx, designation.ID)
	if err != nil {
		t.Fatalf("DeleteDesignation failed: %v", err)
	}

	_, err = designationRepo.GetDesignationByID(ctx, designation.ID)
	if err == nil {
		t.Error("expected error after delete (record should not exist)")
	}
}
