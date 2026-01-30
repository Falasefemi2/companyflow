package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupTestPool(t *testing.T) *pgxpool.Pool {
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
	}
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			t.Logf("✓ Loaded .env from: %s", path)
			break
		}
	}

	connStr := os.Getenv("TEST_DATABASE_URL")

	if connStr == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		sslMode := os.Getenv("DB_SSLMODE")

		if host == "" || user == "" || dbName == "" {
			t.Fatal("missing database configuration. Set TEST_DATABASE_URL or DB_HOST, DB_USER, DB_NAME in .env file")
		}

		if port == "" {
			port = "5432"
		}
		if sslMode == "" {
			sslMode = "require"
		}

		connStr = fmt.Sprintf(
			"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbName, sslMode,
		)
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	return pool
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	pool := setupTestPool(t)
	t.Cleanup(func() {
		pool.Close()
	})
	return pool
}

func setupEmployeeRepository(t *testing.T) *EmployeeRepository {
	pool := setupTestDB(t)
	return NewEmployeeRepository(pool)
}

func setupDepartmentRepository(t *testing.T) *DepartmentRepository {
	pool := setupTestDB(t)
	return NewDepartmentRepository(pool)
}

// setupLevelRepository creates a LevelRepository with test database
func setupLevelRepository(t *testing.T) *LevelRepository {
	pool := setupTestDB(t)
	return NewLevelRepository(pool)
}

// setupDesignationRepository creates a DesignationRepository with test database
func setupDesignationRepository(t *testing.T) *DesignationRepository {
	pool := setupTestDB(t)
	return NewDesignationRepository(pool)
}

// cleanupEmployeeTestData removes test employees by company_id
func cleanupEmployeeTestData(ctx context.Context, pool *pgxpool.Pool, companyID string) error {
	query := `DELETE FROM employees WHERE company_id = $1`
	_, err := pool.Exec(ctx, query, companyID)
	return err
}

// cleanupDepartmentTestData removes test departments by company_id
func cleanupDepartmentTestData(ctx context.Context, pool *pgxpool.Pool, companyID string) error {
	query := `DELETE FROM departments WHERE company_id = $1`
	_, err := pool.Exec(ctx, query, companyID)
	return err
}

// cleanupLevelTestData removes test levels by company_id
func cleanupLevelTestData(ctx context.Context, pool *pgxpool.Pool, companyID string) error {
	query := `DELETE FROM levels WHERE company_id = $1`
	_, err := pool.Exec(ctx, query, companyID)
	return err
}

// cleanupDesignationTestData removes test designations by company_id
func cleanupDesignationTestData(ctx context.Context, pool *pgxpool.Pool, companyID string) error {
	query := `DELETE FROM designations WHERE company_id = $1`
	_, err := pool.Exec(ctx, query, companyID)
	return err
}

// cleanupByEmailPattern removes test data by email pattern (useful for targeted cleanup)
func cleanupByEmailPattern(ctx context.Context, pool *pgxpool.Pool, emailPattern string) error {
	query := `DELETE FROM employees WHERE email LIKE $1`
	_, err := pool.Exec(ctx, query, emailPattern)
	return err
}

// getTestPool returns the raw pool (use sparingly - prefer setupTestDB)
func getTestPool(t *testing.T) *pgxpool.Pool {
	return setupTestDB(t)
}
