package main

import (
	"fmt"
	"os"

	"github.com/arigatosimarmata/artech-user/internal/config"
	"github.com/arigatosimarmata/artech-user/internal/migration"
	"github.com/spf13/cobra"
)

var (
	migrationDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration tool",
		Long:  "Database migration tool for Artech User Service",
	}

	// Add migration directory flag
	rootCmd.PersistentFlags().StringVarP(&migrationDir, "dir", "d", "migration/sql", "Migration files directory")

	// Up command
	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE:  runUp,
	}

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE:  runStatus,
	}

	// Down command
	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		RunE:  runDown,
	}

	rootCmd.AddCommand(upCmd, statusCmd, downCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runUp(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log, err := config.InitLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Sync()

	// Initialize database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer config.CloseDatabase(db)

	// Create migrator
	migrator := migration.NewMigrator(db, log)

	// Run migrations
	if err := migrator.Up(migrationDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Migrations completed successfully")
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log, err := config.InitLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Sync()

	// Initialize database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer config.CloseDatabase(db)

	// Create migrator
	migrator := migration.NewMigrator(db, log)

	// Show status
	return migrator.Status(migrationDir)
}

func runDown(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log, err := config.InitLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Sync()

	// Initialize database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer config.CloseDatabase(db)

	// Create migrator
	migrator := migration.NewMigrator(db, log)

	// Rollback migration
	if err := migrator.Down(); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	fmt.Println("Migration rolled back successfully")
	return nil
}
