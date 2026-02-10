package config

import (
	"fmt"
	"os"

	"github.com/arigatosimarmata/artech-user/pkg/logger"
)

// InitLogger initializes the logger
func InitLogger(cfg *Config) (logger.Logger, error) {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create logs directory: %w", err)
		}
	}

	log, err := logger.NewLogger(
		cfg.Logger.Level,
		cfg.Logger.FilePath,
		cfg.Logger.MaxSize,
		cfg.Logger.MaxBackups,
		cfg.Logger.MaxAge,
		cfg.Logger.Compress,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return log, nil
}
