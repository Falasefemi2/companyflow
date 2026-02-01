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

func ptrFloat(f float64) *float64 {
	return &f
}

func TestLevelRepository_CreateLevel(t *testing.T) {
	repo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupLevelTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level := &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Senior Developer-%d", time.Now().UnixNano()),
		HierarchyLevel: 1,
		MinSalary:      ptrFloat(5000000.00),
		MaxSalary:      ptrFloat(15000000.00),
		Description:    "Senoir Level",
	}

	result, err := repo.CreateLevel(ctx, level)
	if err != nil {
		t.Fatalf("CreateLevel failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected level, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.Name != level.Name {
		t.Errorf("expected name %s, got %s", level.Name, result.Name)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}

	if result.MinSalary == nil {
		t.Error("expected MinSalary to be set, got nil")
	} else if *result.MinSalary != 5000000.00 {
		t.Errorf("expected MinSalary 5000000.00, got %v", *result.MinSalary)
	}

	if result.MaxSalary == nil {
		t.Error("expected MaxSalary to be set, got nil")
	} else if *result.MaxSalary != 15000000.00 {
		t.Errorf("expected MaxSalary 15000000.00, got %v", *result.MaxSalary)
	}

	if result.HierarchyLevel != level.HierarchyLevel {
		t.Errorf("expected HierarchyLevel %d, got %d", level.HierarchyLevel, result.HierarchyLevel)
	}
}

func TestLevelRepository_GetLevelById(t *testing.T) {
	repo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupLevelTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	level := &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Senior Developer-%d", time.Now().UnixNano()),
		HierarchyLevel: 1,
		MinSalary:      ptrFloat(5000000.00),
		MaxSalary:      ptrFloat(15000000.00),
		Description:    "Senoir Level",
	}

	created, err := repo.CreateLevel(ctx, level)
	if err != nil {
		t.Fatalf("creation failed: %v", err)
	}

	result, err := repo.GetLevelByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetLevelByid failed: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("expected %v, got %v", created.ID, result.ID)
	}

	if result.Name != created.Name {
		t.Errorf("expected name %s, got %s", created.Name, result.Name)
	}

	if result.MinSalary == nil {
		t.Error("expected MinSalary to be set, got nil")
	} else if *result.MinSalary != 5000000.00 {
		t.Errorf("expected MinSalary 5000000.00, got %v", *result.MinSalary)
	}

	if result.MaxSalary == nil {
		t.Error("expected MaxSalary to be set, got nil")
	} else if *result.MaxSalary != 15000000.00 {
		t.Errorf("expected MaxSalary 15000000.00, got %v", *result.MaxSalary)
	}

	if result.HierarchyLevel != level.HierarchyLevel {
		t.Errorf("expected HierarchyLevel %d, got %d", level.HierarchyLevel, result.HierarchyLevel)
	}
}

func TestLevelRepository_GetLevelList(t *testing.T) {
	repo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupLevelTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	for i := 1; i <= 5; i++ {
		_, err := repo.CreateLevel(ctx, &models.Level{
			CompanyID:      companyID,
			Name:           fmt.Sprintf("Level%d-%d", i, time.Now().UnixNano()),
			HierarchyLevel: i,
			MinSalary:      ptrFloat(float64(1000000 * i)),
			MaxSalary:      ptrFloat(float64(5000000 * i)),
			Description:    fmt.Sprintf("Level %d", i),
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	req := &dto.LevelListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 2,
		},
	}

	result, err := repo.GetLevelList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetLevelList failed: %v", err)
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

func TestLevelRepository_GetLevelList_SearchByName(t *testing.T) {
	repo := setupLevelRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupLevelTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	_, err := repo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Executive Manager-%d", time.Now().UnixNano()),
		HierarchyLevel: 5,
		MinSalary:      ptrFloat(10000000.00),
		MaxSalary:      ptrFloat(20000000.00),
		Description:    "Executive level",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := &dto.LevelListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		Search: "Executive",
	}

	result, err := repo.GetLevelList(ctx, companyID, req)
	if err != nil {
		t.Fatalf("GetLevelList failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected results for Executive search")
	}
}

func TestLevelRepository_UpdateLevel(t *testing.T) {
	repo := setupLevelRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	level, err := repo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Original-%d", time.Now().UnixNano()),
		HierarchyLevel: 2,
		MinSalary:      ptrFloat(3000000.00),
		MaxSalary:      ptrFloat(8000000.00),
		Description:    "Original description",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newName := fmt.Sprintf("Updated-%d", time.Now().UnixNano())
	newMinSalary := 4000000.00
	newMaxSalary := 10000000.00

	updated, err := repo.UpdateLevel(ctx, level.ID, &models.Level{
		Name:      newName,
		MinSalary: &newMinSalary,
		MaxSalary: &newMaxSalary,
	})
	if err != nil {
		t.Fatalf("UpdateLevel failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}

	if updated.MinSalary == nil || *updated.MinSalary != newMinSalary {
		t.Errorf("expected MinSalary %v, got %v", newMinSalary, updated.MinSalary)
	}

	if updated.MaxSalary == nil || *updated.MaxSalary != newMaxSalary {
		t.Errorf("expected MaxSalary %v, got %v", newMaxSalary, updated.MaxSalary)
	}

	fetched, err := repo.GetLevelByID(ctx, level.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if fetched.Name != newName {
		t.Errorf("persisted name should be %s, got %s", newName, fetched.Name)
	}
}

func TestLevelRepository_DeleteLevel(t *testing.T) {
	repo := setupLevelRepository(t)
	ctx := context.Background()

	companyID := uuid.MustParse(testCompanyID)

	level, err := repo.CreateLevel(ctx, &models.Level{
		CompanyID:      companyID,
		Name:           fmt.Sprintf("Delete-%d", time.Now().UnixNano()),
		HierarchyLevel: 3,
		Description:    "To be deleted",
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = repo.DeleteLevel(ctx, level.ID)
	if err != nil {
		t.Fatalf("DeleteLevel failed: %v", err)
	}

	_, err = repo.GetLevelByID(ctx, level.ID)
	if err == nil {
		t.Error("expected error after delete (record should not exist)")
	}
}
