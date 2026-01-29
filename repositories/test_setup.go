package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// setupTestPool creates a database connection pool for testing
// It loads from TEST_DATABASE_URL first, then falls back to individual env vars
func setupTestPool(t *testing.T) *pgxpool.Pool {
	// Try to load .env from multiple paths (for different working directories)
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

	// Try to get TEST_DATABASE_URL first (Neon full URL)
	connStr := os.Getenv("TEST_DATABASE_URL")

	// If not set, fall back to building from individual variables
	if connStr == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		sslMode := os.Getenv("DB_SSLMODE")

		// Validate required variables
		if host == "" || user == "" || dbName == "" {
			t.Fatal("missing database configuration. Set TEST_DATABASE_URL or DB_HOST, DB_USER, DB_NAME in .env file")
		}

		// Set defaults if needed
		if port == "" {
			port = "5432"
		}
		if sslMode == "" {
			sslMode = "require"
		}

		// Build connection string
		connStr = fmt.Sprintf(
			"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbName, sslMode,
		)
	}

	// Create connection pool
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Test the connection
	err = pool.Ping(context.Background())
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	return pool
}

// setupTestRepository creates an EmployeeRepository for testing
func setupTestRepository(t *testing.T) *EmployeeRepository {
	pool := setupTestPool(t)

	// Register cleanup - close pool after test completes
	t.Cleanup(func() {
		pool.Close()
	})

	return NewEmployeeRepository(pool)
}

// cleanupTestData deletes test employees by email pattern
// Call this inside your test if you want to clean up specific test data
// cleanupEmployeeTestData removes test employees before running a test
func cleanupEmployeeTestData(ctx context.Context, pool *pgxpool.Pool, companyID string) error {
	query := `DELETE FROM employees WHERE company_id = $1`
	_, err := pool.Exec(ctx, query, companyID)
	return err
}
