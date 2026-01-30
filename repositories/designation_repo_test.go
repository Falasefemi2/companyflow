package repositories

// import (
// 	"context"
// 	"fmt"
// 	"reflect"
// 	"testing"
// 	"time"
//
// 	"github.com/google/uuid"
//
// 	"github.com/falasefemi2/companyflowlow/models"
// )
//
// func TestDesignationRepository_CreateDesignation(t *testing.T) {
// 	repo := setupDesignationRepository(t)
// 	pool := setupTestDB(t)
// 	ctx := context.Background()
//
// 	companyID := uuid.MustParse(testCompanyID)
//
// 	if err := cleanupDesignationTestData(ctx, pool, companyID.String()); err != nil {
// 		t.Fatalf("cleanup failed: %v", err)
// 	}
//
// 	designation := &models.Designation{
// 		CompanyID:   companyID,
// 		Name:        fmt.Sprintf("Designation%d", time.Now().UnixNano()),
// 		Description: "Designation description",
// 		LevelID:     *uuid.UUID,
// 	}
// }
