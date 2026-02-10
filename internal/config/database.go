package config

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// InitDatabase initializes database connection
func InitDatabase(cfg *Config) (*sqlx.DB, error) {
	dsn := cfg.GetDSN()
	
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Ping database to verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *sqlx.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
