package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	sort.Strings(files)

	for _, file := range files {
		log.Printf("Applying migration: %s", file)

		migrationSQL, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		log.Printf("Successfully applied: %s", file)
	}

	log.Println("All migrations completed successfully")
	return nil
}
