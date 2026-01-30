package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/models"
)

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
		Name:           "Senoir Developer",
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
}

func ptrFloat(f float64) *float64 {
	return &f
}
