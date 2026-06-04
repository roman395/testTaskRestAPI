package database

import (
	"api/config"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(cfg *config.Config) (*sqlx.DB, error) {
	log.Println("[DATABASE] Connecting to database...")
	log.Printf("[DATABASE] Host: %s, Port: %d, Database: %s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL())
	if err != nil {
		log.Printf("[DATABASE] Connection failed: %v", err)
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime())

	log.Printf("[DATABASE] Connection pool configured: max_open=%d, max_idle=%d, max_lifetime=%v",
		cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.ConnMaxLifetime())

	// Test connection
	if err := db.Ping(); err != nil {
		log.Printf("[DATABASE] Ping failed: %v", err)
		return nil, err
	}

	log.Println("[DATABASE] Connection established successfully")
	return db, nil
}
