package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arigatosimarmata/artech-user/pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Migrator handles database migrations
type Migrator struct {
	db     *sqlx.DB
	logger logger.Logger
}

// NewMigrator creates a new migrator
func NewMigrator(db *sqlx.DB, logger logger.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// Initialize creates the migrations table if it doesn't exist
func (m *Migrator) Initialize() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	m.logger.Info("Migrations table initialized")
	return nil
}

// LoadMigrations loads migration files from a directory
func (m *Migrator) LoadMigrations(migrationDir string) ([]Migration, error) {
	var migrations []Migration

	// Read migration files
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Parse version from filename (e.g., 001_create_users_table.sql)
		parts := strings.SplitN(file.Name(), "_", 2)
		if len(parts) < 2 {
			m.logger.Warn("Skipping invalid migration file", zap.String("file", file.Name()))
			continue
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			m.logger.Warn("Failed to parse migration version", zap.String("file", file.Name()), zap.Error(err))
			continue
		}

		// Read SQL content
		filePath := filepath.Join(migrationDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    strings.TrimSuffix(file.Name(), ".sql"),
			SQL:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetAppliedMigrations gets the list of applied migrations
func (m *Migrator) GetAppliedMigrations() (map[int]bool, error) {
	query := "SELECT version FROM schema_migrations"
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// Up applies all pending migrations
func (m *Migrator) Up(migrationDir string) error {
	// Initialize migrations table
	if err := m.Initialize(); err != nil {
		return err
	}

	// Load migrations
	migrations, err := m.LoadMigrations(migrationDir)
	if err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if applied[migration.Version] {
			m.logger.Info("Migration already applied", zap.Int("version", migration.Version), zap.String("name", migration.Name))
			continue
		}

		m.logger.Info("Applying migration", zap.Int("version", migration.Version), zap.String("name", migration.Name))

		// Start transaction
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// Execute migration
		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
		}

		// Record migration
		recordQuery := "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)"
		if _, err := tx.Exec(recordQuery, migration.Version, migration.Name, time.Now()); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		m.logger.Info("Migration applied successfully", zap.Int("version", migration.Version))
	}

	m.logger.Info("All migrations applied successfully")
	return nil
}

// Status shows the status of migrations
func (m *Migrator) Status(migrationDir string) error {
	// Initialize migrations table
	if err := m.Initialize(); err != nil {
		return err
	}

	// Load migrations
	migrations, err := m.LoadMigrations(migrationDir)
	if err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("================")

	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	for _, migration := range migrations {
		status := "Pending"
		if applied[migration.Version] {
			status = "Applied"
		}
		fmt.Printf("[%s] %03d_%s\n", status, migration.Version, migration.Name)
	}

	fmt.Println()
	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	// Get the last applied migration
	query := "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1"
	
	var version int
	err := m.db.Get(&version, query)
	if err != nil {
		if err == sql.ErrNoRows {
			m.logger.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	// Delete migration record
	deleteQuery := "DELETE FROM schema_migrations WHERE version = ?"
	if _, err := m.db.Exec(deleteQuery, version); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", version, err)
	}

	m.logger.Info("Migration rolled back", zap.Int("version", version))
	return nil
}
