package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(db *pgxpool.Pool) error {
	ctx := context.Background()

	createMigrationsTable := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if err := db.QueryRow(ctx, createMigrationsTable).Scan(); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			fmt.Println("Note: schema_migrations table may already exist")
		}
	}

	migrationsDir := "database/migration"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles) // Sort by name (000_, 001_, 002_, etc.)

	for _, filename := range sqlFiles {
		var count int
		err := db.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE name = $1", filename).Scan(&count)
		if err != nil {
			return fmt.Errorf("error checking migration status: %w", err)
		}

		if count > 0 {
			fmt.Printf("✓ Migration already ran: %s\n", filename)
			continue
		}

		filePath := filepath.Join(migrationsDir, filename)
		sqlBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %w", filename, err)
		}

		sql := string(sqlBytes)

		if _, err := db.Exec(ctx, sql); err != nil {
			return fmt.Errorf("error executing migration %s: %w", filename, err)
		}

		if _, err := db.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", filename); err != nil {
			return fmt.Errorf("error recording migration %s: %w", filename, err)
		}

		fmt.Printf("✓ Migration executed: %s\n", filename)
	}

	fmt.Println("\n✓ All migrations completed successfully!")
	return nil
}
