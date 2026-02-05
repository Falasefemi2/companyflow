package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/google/uuid"
)

func ptr[T any](v T) *T {
	return &v
}

func TestRoleRepository_CreateRole(t *testing.T) {
	repo := setupRoleRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()
	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupRoleTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	role := &models.Role{
		CompanyID:        &companyID,
		Name:             fmt.Sprintf("Driver%d", time.Now().UnixNano()),
		Description:      ptr("Transportation"),
		IsSystemRole:     false,
		PermissionsCache: []string{"users.read", "trips.view"},
	}

	result, err := repo.CreateRole(ctx, role)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected role, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	if result.Name != role.Name {
		t.Errorf("expected name %s, got %s", role.Name, result.Name)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestRoleRepository_GetRoleByID(t *testing.T) {
	repo := setupRoleRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()
	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupRoleTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	role := &models.Role{
		CompanyID:        &companyID,
		Name:             fmt.Sprintf("Driver%d", time.Now().UnixNano()),
		Description:      ptr("Transportation"),
		IsSystemRole:     false,
		PermissionsCache: []string{"users.read", "trips.view"},
	}

	created, err := repo.CreateRole(ctx, role)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result, err := repo.GetRoleByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetRoleByID failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected role, got nil")
	}

	if result.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, result.ID)
	}

	if result.Name != created.Name {
		t.Errorf("expected name %s, got %s", created.Name, result.Name)
	}

	if result.CompanyID == nil || *result.CompanyID != *created.CompanyID {
		t.Error("company ID mismatch")
	}

	if len(result.PermissionsCache) != len(created.PermissionsCache) {
		t.Errorf("expected %d permissions, got %d",
			len(created.PermissionsCache),
			len(result.PermissionsCache),
		)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

// TestRoleRepository_GetRoleList tests listing roles with pagination and search
func TestRoleRepository_GetRoleList(t *testing.T) {
	repo := setupRoleRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()
	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupRoleTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Create test roles
	roles := []string{"Driver", "Manager", "Admin", "Viewer"}
	createdRoles := make([]*models.Role, len(roles))

	for i, name := range roles {
		role := &models.Role{
			CompanyID:        &companyID,
			Name:             fmt.Sprintf("%s_%d", name, time.Now().UnixNano()+int64(i)),
			Description:      ptr(fmt.Sprintf("%s role", name)),
			IsSystemRole:     false,
			PermissionsCache: []string{"read", "write"},
		}

		created, err := repo.CreateRole(ctx, role)
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}
		createdRoles[i] = created
	}

	t.Run("get all roles with pagination", func(t *testing.T) {
		listReq := &dto.RoleListRequest{
			Page:     1,
			PageSize: 10,
		}

		result, err := repo.GetRoleList(ctx, companyID, listReq)
		if err != nil {
			t.Fatalf("GetRoleList failed: %v", err)
		}

		if result == nil {
			t.Fatal("expected response, got nil")
		}

		if len(result.Data) != len(roles) {
			t.Errorf("expected %d roles, got %d", len(roles), len(result.Data))
		}

		if result.Total != int64(len(roles)) {
			t.Errorf("expected total %d, got %d", len(roles), result.Total)
		}

		if result.Page != 1 {
			t.Errorf("expected page 1, got %d", result.Page)
		}

		if result.HasNext {
			t.Error("expected HasNext to be false")
		}

		if result.HasPrev {
			t.Error("expected HasPrev to be false")
		}
	})

	t.Run("search by role name", func(t *testing.T) {
		// Search for "Driver" in the name
		searchTerm := createdRoles[0].Name[:6] // "Driver"
		listReq := &dto.RoleListRequest{
			Page:     1,
			PageSize: 10,
			Search:   searchTerm,
		}

		result, err := repo.GetRoleList(ctx, companyID, listReq)
		if err != nil {
			t.Fatalf("GetRoleList with search failed: %v", err)
		}

		if len(result.Data) == 0 {
			t.Errorf("expected to find role with search term %s", searchTerm)
		}

		// Verify the search result contains the search term
		for _, role := range result.Data {
			if !contains(role.Name, searchTerm) {
				t.Errorf("expected role name to contain %s, got %s", searchTerm, role.Name)
			}
		}
	})

	t.Run("pagination - second page", func(t *testing.T) {
		listReq := &dto.RoleListRequest{
			Page:     2,
			PageSize: 2,
		}

		result, err := repo.GetRoleList(ctx, companyID, listReq)
		if err != nil {
			t.Fatalf("GetRoleList pagination failed: %v", err)
		}

		if result.Page != 2 {
			t.Errorf("expected page 2, got %d", result.Page)
		}

		if len(result.Data) != 2 {
			t.Errorf("expected 2 roles on page 2, got %d", len(result.Data))
		}

		if result.HasPrev != true {
			t.Error("expected HasPrev to be true on page 2")
		}
	})
}

// TestRoleRepository_UpdateRole tests updating role fields
func TestRoleRepository_UpdateRole(t *testing.T) {
	repo := setupRoleRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()
	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupRoleTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Create a role to update
	originalRole := &models.Role{
		CompanyID:        &companyID,
		Name:             fmt.Sprintf("OriginalRole%d", time.Now().UnixNano()),
		Description:      ptr("Original description"),
		IsSystemRole:     false,
		PermissionsCache: []string{"read"},
	}

	created, err := repo.CreateRole(ctx, originalRole)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("update name only", func(t *testing.T) {
		updateRole := &models.Role{
			Name:        "UpdatedRoleName",
			Description: nil, // nil means don't update
		}

		result, err := repo.UpdateRole(ctx, created.ID, updateRole)
		if err != nil {
			t.Fatalf("UpdateRole failed: %v", err)
		}

		if result.Name != "UpdatedRoleName" {
			t.Errorf("expected name to be updated to UpdatedRoleName, got %s", result.Name)
		}

		if *result.Description != "Original description" {
			t.Errorf("expected description to remain unchanged, got %s", *result.Description)
		}

		if result.UpdatedAt.Before(created.UpdatedAt) {
			t.Error("expected UpdatedAt to be newer")
		}
	})

	t.Run("update description only", func(t *testing.T) {
		updateRole := &models.Role{
			Name:             "",
			Description:      ptr("New description"),
			PermissionsCache: nil,
		}

		result, err := repo.UpdateRole(ctx, created.ID, updateRole)
		if err != nil {
			t.Fatalf("UpdateRole failed: %v", err)
		}

		if result.Name != "UpdatedRoleName" {
			t.Errorf("expected name to remain UpdatedRoleName, got %s", result.Name)
		}

		if *result.Description != "New description" {
			t.Errorf("expected description to be New description, got %s", *result.Description)
		}
	})

	t.Run("update is_system_role remains unchanged", func(t *testing.T) {
		// Try to update is_system_role (this should NOT change in DB because we don't update it)
		updateRole := &models.Role{
			Name:         "",
			IsSystemRole: true, // attempt to change this
		}

		result, err := repo.UpdateRole(ctx, created.ID, updateRole)
		if err != nil {
			t.Fatalf("UpdateRole failed: %v", err)
		}

		// Verify is_system_role was NOT changed
		if result.IsSystemRole != false {
			t.Error("expected is_system_role to remain false (protected field)")
		}
	})
}

// TestRoleRepository_DeleteRole tests hard delete functionality
func TestRoleRepository_DeleteRole(t *testing.T) {
	repo := setupRoleRepository(t)
	pool := setupTestDB(t)
	ctx := context.Background()
	companyID := uuid.MustParse(testCompanyID)

	if err := cleanupRoleTestData(ctx, pool, companyID.String()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Create a role to delete
	role := &models.Role{
		CompanyID:        &companyID,
		Name:             fmt.Sprintf("DeleteMe%d", time.Now().UnixNano()),
		Description:      ptr("To be deleted"),
		IsSystemRole:     false,
		PermissionsCache: []string{"read"},
	}

	created, err := repo.CreateRole(ctx, role)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("hard delete removes role from database", func(t *testing.T) {
		// Delete the role
		err := repo.DeleteRole(ctx, created.ID)
		if err != nil {
			t.Fatalf("DeleteRole failed: %v", err)
		}

		// Verify role is actually deleted by trying to get it
		_, err = repo.GetRoleByID(ctx, created.ID)
		if err == nil {
			t.Error("expected error when getting deleted role, but got none")
		}
	})

	t.Run("delete non-existent role returns error", func(t *testing.T) {
		fakeID := uuid.New()
		err := repo.DeleteRole(ctx, fakeID)

		if err == nil {
			t.Error("expected error when deleting non-existent role, got nil")
		}

		if err.Error() != "role not found" {
			t.Errorf("expected 'role not found' error, got %v", err)
		}
	})
}

// Helper function to check if a string contains a substring (case-sensitive)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
