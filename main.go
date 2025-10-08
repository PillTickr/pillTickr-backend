package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"pillTickr-backend/crypto"
	"pillTickr-backend/db"
	"pillTickr-backend/middleware"
	"pillTickr-backend/routes"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using system environment variables", "error", err)
	}

	// Initialize database with error handling
	if err := db.Init("records.db"); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Set encryption key with validation
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		slog.Error("ENCRYPTION_KEY environment variable is required")
		os.Exit(1)
	}

	if err := crypto.SetKey([]byte(key)); err != nil {
		slog.Error("Failed to set encryption key", "error", err)
		os.Exit(1)
	}

	slog.Info("Application initialized successfully")
}

func main() {
	slog.Info("Starting PillTickr Server...")

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := gin.Default()

	// Add middleware in order
	server.Use(middleware.ErrorHandler())    // Panic recovery
	server.Use(middleware.SecurityHeaders()) // Security headers
	// server.Use(middleware.RateLimiter())       // Rate limiting

	// this provides logging HTTP Requests
	server.Use(middleware.RequestLogger())          // Request logging
	server.Use(middleware.ValidationErrorHandler()) // Validation error handling

	// Log the Origin header for every request
	server.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = c.Request.Header.Get("Referer")
			if origin == "" {
				origin = "<none>"
			}
		}

		slog.Info("Incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"origin", origin,
		)
		c.Next()
	})

	// Add CORS middleware
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	prefix := "/api"
	apiGroup := server.Group(prefix)

	apiRoutes := routes.NewRoutes()
	routes.AttachRoutes(apiGroup, apiRoutes)

	// Health check endpoint
	server.GET("/health", func(c *gin.Context) {
		if err := db.HealthCheck(); err != nil {
			slog.Error("Health check failed", "error", err)
			c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// A nice get started message
	server.GET("/", func(c *gin.Context) {
		c.String(200, "Welcome to the PillTickr API! Visit /api/docs for API documentation.")
	})

	// Graceful shutdown setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090" // fallback default
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Server running at http://localhost:" + port)
		if err := server.Run(":" + port); err != nil {
			slog.Error("Server failed to start", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Cleanup
	slog.Info("Shutting down server...")
	if err := db.Close(); err != nil {
		slog.Error("Failed to close database", "error", err)
	}
	slog.Info("Server shutdown complete")
}
