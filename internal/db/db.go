package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// RunMigrations executes all .sql files from the given directory in order.
// It tracks applied migrations in a schema_migrations table.
func RunMigrations(db *sqlx.DB, migrationsDir string) error {
	// Ensure tracking table exists
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Get already-applied migrations
	var applied []string
	if err := db.Select(&applied, `SELECT filename FROM schema_migrations ORDER BY filename`); err != nil {
		return fmt.Errorf("query schema_migrations: %w", err)
	}
	appliedSet := make(map[string]bool)
	for _, f := range applied {
		appliedSet[f] = true
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", migrationsDir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		if appliedSet[f] {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		log.Printf("applying migration: %s", f)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, f); err != nil {
			return fmt.Errorf("record %s: %w", f, err)
		}
	}

	return nil
}
