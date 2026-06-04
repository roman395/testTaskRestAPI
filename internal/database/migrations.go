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
	log.Println("Checking for pending migrations...")

	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		log.Println("Migrations directory not found, skipping migrations")
		return nil
	}

	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		log.Printf("Failed to read migrations: %v", err)
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	if len(files) == 0 {
		log.Println("No migration files found")
		return nil
	}

	log.Printf("Found %d migration files", len(files))

	sort.Strings(files)

	for i, file := range files {
		log.Printf("Applying migration %d/%d: %s", i+1, len(files), filepath.Base(file))

		migrationSQL, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Failed to read migration file %s: %v", file, err)
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			log.Printf("Failed to execute migration %s: %v", file, err)
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		log.Printf("Successfully applied migration: %s", filepath.Base(file))
	}

	log.Println("All migrations completed successfully")
	return nil
}
