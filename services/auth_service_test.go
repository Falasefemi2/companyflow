package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

func setupAuthService(t *testing.T) (*AuthService, *repositories.EmployeeRepository, *repositories.CompanyRepository) {
	pool := setupTestDB(t)
	employeeRepo := repositories.NewEmployeeRepository(pool)
	companyRepo := repositories.NewCompanyRepository(pool)
	authService := NewAuthService(employeeRepo, companyRepo)
	return authService, employeeRepo, companyRepo
}

func createTestCompanyForAuth(t *testing.T, companyRepo *repositories.CompanyRepository) uuid.UUID {
	ctx := context.Background()

	company := &dto.CreateCompanyRequest{
		Name: "Auth Test Company",
		Slug: fmt.Sprintf("auth-test-%d", time.Now().UnixNano()),
	}

	companyService := &CompanyService{companyRepo: companyRepo}
	created, err := companyService.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("failed to create test company: %v", err)
	}

	return uuid.MustParse(created.ID)
}

func createTestRoleForAuth(t *testing.T, pool interface{}, companyID uuid.UUID, roleName string) uuid.UUID {
	ctx := context.Background()
	roleID := uuid.New()

	// Type assert the pool to get the underlying connection
	if dbPool, ok := pool.(interface {
		Exec(context.Context, string, ...interface{}) (interface{}, error)
	}); ok {
		query := `
			INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err := dbPool.Exec(ctx, query, roleID, companyID, roleName, "Test role")
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}
	}

	return roleID
}

func createTestEmployeeForAuth(t *testing.T, employeeRepo *repositories.EmployeeRepository, companyID uuid.UUID, email string, password string, roleName string) uuid.UUID {
	ctx := context.Background()

	// Create role first
	pool := setupTestDB(t)
	roleID := createTestRoleForAuth(t, pool, companyID, roleName)

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	employeeID := uuid.New()

	query := `
		INSERT INTO employees (id, company_id, email, first_name, last_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	// We need direct pool access for this
	if dbPool, ok := interface{}(employeeRepo).(interface {
		Exec(context.Context, string, ...interface{}) error
	}); ok {
		err = dbPool.Exec(ctx, query, employeeID, companyID, email, "Test", "Employee", hashedPassword, roleID, "active")
		if err != nil {
			t.Fatalf("failed to create test employee: %v", err)
		}
	}

	return employeeID
}

func TestAuthService_Login_Success(t *testing.T) {
	authService, _, companyRepo := setupAuthService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	// Create test company
	companyID := createTestCompanyForAuth(t, companyRepo)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	// Create Super Admin role
	roleID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roleID, companyID, "Super Admin", "Admin role")
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create test employee
	email := fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano())
	password := "TestPassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	employeeID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO employees (id, company_id, email, first_name, last_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, employeeID, companyID, email, "Test", "Admin", hashedPassword, roleID, "active")
	if err != nil {
		t.Fatalf("failed to create employee: %v", err)
	}

	// Test login
	req := &dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	result, err := authService.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected login response, got nil")
	}

	if result.Token == "" {
		t.Error("expected token to be set")
	}

	if result.Role != "Super Admin" {
		t.Errorf("expected role 'Super Admin', got %s", result.Role)
	}

	if result.Employee == nil {
		t.Error("expected employee to be set")
	}

	if result.Company == nil {
		t.Error("expected company to be set")
	}
}

func TestAuthService_Login_HRManager(t *testing.T) {
	authService, _, companyRepo := setupAuthService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	// Create test company
	companyID := createTestCompanyForAuth(t, companyRepo)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	// Create HR Manager role
	roleID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roleID, companyID, "HR Manager", "HR role")
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create test employee
	email := fmt.Sprintf("hr-%d@example.com", time.Now().UnixNano())
	password := "TestPassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	employeeID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO employees (id, company_id, email, first_name, last_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, employeeID, companyID, email, "Test", "HR", hashedPassword, roleID, "active")
	if err != nil {
		t.Fatalf("failed to create employee: %v", err)
	}

	// Test login
	req := &dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	result, err := authService.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.Role != "HR Manager" {
		t.Errorf("expected role 'HR Manager', got %s", result.Role)
	}
}

func TestAuthService_Login_NilRequest(t *testing.T) {
	authService, _, _ := setupAuthService(t)
	ctx := context.Background()

	result, err := authService.Login(ctx, nil)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestAuthService_Login_EmptyEmail(t *testing.T) {
	authService, _, _ := setupAuthService(t)
	ctx := context.Background()

	req := &dto.LoginRequest{
		Email:    "",
		Password: "password",
	}

	result, err := authService.Login(ctx, req)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestAuthService_Login_EmptyPassword(t *testing.T) {
	authService, _, _ := setupAuthService(t)
	ctx := context.Background()

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "",
	}

	result, err := authService.Login(ctx, req)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	authService, _, _ := setupAuthService(t)
	ctx := context.Background()

	req := &dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password",
	}

	result, err := authService.Login(ctx, req)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	authService, _, companyRepo := setupAuthService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	// Create test company
	companyID := createTestCompanyForAuth(t, companyRepo)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	// Create Super Admin role
	roleID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roleID, companyID, "Super Admin", "Admin role")
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create test employee
	email := fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano())
	password := "CorrectPassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	employeeID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO employees (id, company_id, email, first_name, last_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, employeeID, companyID, email, "Test", "Admin", hashedPassword, roleID, "active")
	if err != nil {
		t.Fatalf("failed to create employee: %v", err)
	}

	// Test login with wrong password
	req := &dto.LoginRequest{
		Email:    email,
		Password: "WrongPassword123",
	}

	result, err := authService.Login(ctx, req)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestAuthService_Login_InsufficientRole(t *testing.T) {
	authService, _, companyRepo := setupAuthService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	// Create test company
	companyID := createTestCompanyForAuth(t, companyRepo)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	// Create regular Employee role (not Super Admin or HR Manager)
	roleID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roleID, companyID, "Employee", "Regular employee")
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create test employee
	email := fmt.Sprintf("employee-%d@example.com", time.Now().UnixNano())
	password := "TestPassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	employeeID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO employees (id, company_id, email, first_name, last_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, employeeID, companyID, email, "Test", "Employee", hashedPassword, roleID, "active")
	if err != nil {
		t.Fatalf("failed to create employee: %v", err)
	}

	// Test login
	req := &dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	result, err := authService.Login(ctx, req)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	authService, _, companyRepo := setupAuthService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	// Create test company
	companyID := createTestCompanyForAuth(t, companyRepo)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	// Create Super Admin role
	roleID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roleID, companyID, "Super Admin", "Admin role")
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create inactive test employee
	email := fmt.Sprintf("inactive-%d@example.com", time.Now().UnixNano())
	password := "TestPassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	employeeID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO employees (id, company_id, email, first_name, last_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, employeeID, companyID, email, "Test", "Admin", hashedPassword, roleID, "inactive")
	if err != nil {
		t.Fatalf("failed to create employee: %v", err)
	}

	// Test login
	req := &dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	result, err := authService.Login(ctx, req)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestAuthService_Login_TokenGeneration(t *testing.T) {
	authService, _, companyRepo := setupAuthService(t)
	pool := setupTestDB(t)
	ctx := context.Background()

	// Create test company
	companyID := createTestCompanyForAuth(t, companyRepo)
	defer pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)

	// Create Super Admin role
	roleID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roleID, companyID, "Super Admin", "Admin role")
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create test employee
	email := fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano())
	password := "TestPassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	employeeID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO employees (id, company_id, email, first_name, last_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, employeeID, companyID, email, "Test", "Admin", hashedPassword, roleID, "active")
	if err != nil {
		t.Fatalf("failed to create employee: %v", err)
	}

	// Test login
	req := &dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	result, err := authService.Login(ctx, req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Verify token is valid
	claims, err := utils.ValidateToken(result.Token)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims.Subject != employeeID.String() {
		t.Errorf("expected subject %s, got %s", employeeID.String(), claims.Subject)
	}

	if claims.Role != "Super Admin" {
		t.Errorf("expected role 'Super Admin', got %s", claims.Role)
	}

	if claims.CompanyID != companyID.String() {
		t.Errorf("expected company ID %s, got %s", companyID.String(), claims.CompanyID)
	}
}