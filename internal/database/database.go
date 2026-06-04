package database

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sqlx.DB, error) {
	log.Println("Initializing database connection...")

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✅ Database connection established successfully")

	if err := db.Ping(); err != nil {
		log.Printf("Database ping failed: %v", err)
		return nil, err
	}

	log.Println("Database ping successful")

	return db, nil
}
