package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arigatosimarmata/artech-user/handler"
	"github.com/arigatosimarmata/artech-user/internal/config"
	"github.com/arigatosimarmata/artech-user/internal/middleware"
	"github.com/arigatosimarmata/artech-user/pkg/email"
	"github.com/arigatosimarmata/artech-user/pkg/token"
	"github.com/arigatosimarmata/artech-user/repository"
	"github.com/arigatosimarmata/artech-user/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "http",
		Short: "HTTP server",
		Long:  "HTTP server for Artech User Service",
	}

	// Start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start HTTP server",
		RunE:  runServer,
	}

	rootCmd.AddCommand(startCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
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

	log.Info("Starting Artech User Service")

	// Initialize database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		log.Error("Failed to initialize database", zap.Error(err))
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer config.CloseDatabase(db)

	log.Info("Database connected successfully")

	// Initialize JWT manager
	jwtManager := token.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)

	// Initialize email sender
	emailSender := email.NewEmailSender(
		cfg.Email.SMTPHost,
		cfg.Email.SMTPPort,
		cfg.Email.SMTPUsername,
		cfg.Email.SMTPPassword,
		cfg.Email.EmailFrom,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	auditTrailRepo := repository.NewAuditTrailRepository(db)

	// Initialize usecases
	userUsecase := usecase.NewUserUsecase(userRepo, auditTrailRepo, jwtManager, emailSender, log)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userUsecase, log)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandlerMiddleware(log),
		AppName:      cfg.App.Name,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(middleware.LoggingMiddleware(log))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    cfg.App.Name,
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Post("/refresh", userHandler.RefreshToken)
	auth.Post("/forgot-password", userHandler.ForgotPassword)
	auth.Post("/reset-password", userHandler.ResetPassword)

	// Auth routes (protected)
	authProtected := auth.Group("")
	authProtected.Use(middleware.AuthMiddleware(jwtManager))
	authProtected.Post("/logout", userHandler.Logout)

	// User routes (protected)
	user := api.Group("/user")
	user.Use(middleware.AuthMiddleware(jwtManager))
	user.Get("/profile", userHandler.GetProfile)

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info("Server starting", zap.String("address", serverAddr))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(serverAddr); err != nil {
			log.Error("Failed to start server", zap.Error(err))
		}
	}()

	log.Info("Server started successfully", zap.String("address", serverAddr))

	<-quit
	log.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Error("Failed to shutdown server gracefully", zap.Error(err))
		return err
	}

	log.Info("Server stopped")
	return nil
}
