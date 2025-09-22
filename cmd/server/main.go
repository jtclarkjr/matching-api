package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "matching-api/docs"
	"matching-api/internal/database"
	"matching-api/internal/handlers/auth"
	"matching-api/internal/handlers/chat"
	"matching-api/internal/handlers/image"
	"matching-api/internal/handlers/match"
	"matching-api/internal/handlers/notification"
	"matching-api/internal/handlers/user"
	customMiddleware "matching-api/internal/middleware"
	"matching-api/pkg/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Matching API - Matching App Backend
// @version 1.0
// @description A complete Go-based REST API for a dating application with real-time chat, sophisticated matching algorithms, and comprehensive user management.
// @termsOfService https://matching-api.example.com/terms

// @contact.name API Support
// @contact.url https://matching-api.example.com/support
// @contact.email support@matching-api.example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Initialize database (optional - will skip if not configured)
	if err := database.Connect(); err != nil {
		log.Printf("Database connection failed (continuing without DB): %v", err)
	} else {
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("Error closing database: %v", err)
			}
		}()
		// Run migrations if database is connected
		if err := database.RunMigrations(); err != nil {
			log.Printf("Migration failed: %v", err)
		}
	}

	// Create router
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))               // Gzip compression
	r.Use(middleware.Timeout(60 * time.Second)) // Request timeout
	r.Use(middleware.Throttle(100))             // Max 100 concurrent requests
	r.Use(middleware.Heartbeat("/ping"))        // Simple health check

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on your frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize Redis service (optional - will skip if not configured)
	var redisService *services.RedisService
	var err error
	redisService, err = services.NewRedisService()
	if err != nil {
		log.Printf("Redis connection failed (continuing without Redis): %v", err)
	} else {
		defer func() {
			if err := redisService.Close(); err != nil {
				log.Printf("Error closing Redis: %v", err)
			}
		}()
		log.Printf("Redis connected successfully")
	}

	// Add Chi session middleware with Redis support
	if redisService != nil {
		r.Use(customMiddleware.NewSessionMiddleware(redisService))
	} else {
		r.Use(customMiddleware.NewSessionMiddleware(nil)) // Fallback to memory
	}

	// Initialize S3 service
	var s3Service *services.S3Service
	if bucketName := os.Getenv("AWS_S3_BUCKET"); bucketName != "" {
		s3Config := services.S3Config{
			Bucket:  bucketName,
			Region:  getEnvOrDefault("AWS_REGION", "us-east-1"),
			BaseURL: os.Getenv("AWS_S3_BASE_URL"), // Optional CDN URL
		}
		var err error
		s3Service, err = services.NewS3Service(s3Config)
		if err != nil {
			log.Printf("Failed to initialize S3 service: %v", err)
		}
	}

	// Initialize handlers with new organized structure
	authHandler := auth.NewHandler(redisService)
	userHandler := user.NewHandler(s3Service, redisService)
	matchHandler := match.NewHandler(redisService)
	chatHandler := chat.NewHandler(redisService)
	notificationHandler := notification.NewHandler(redisService)
	imageHandler := image.NewHandler(s3Service, redisService)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Add Redis-based rate limiting if available
		if redisService != nil {
			r.Use(customMiddleware.APIRateLimit(redisService))
		}

		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			healthStatus := map[string]string{
				"status":   "ok",
				"database": "disconnected",
				"redis":    "disconnected",
			}

			// Check database connection
			if err := database.HealthCheck(); err == nil {
				healthStatus["database"] = "connected"
			}

			// Check Redis connection
			if redisService != nil {
				if err := redisService.HealthCheck(); err == nil {
					healthStatus["redis"] = "connected"
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
				log.Printf("Error encoding health check response: %v", err)
			}
		})

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			// Add stricter rate limiting for auth endpoints
			if redisService != nil {
				r.Use(customMiddleware.AuthRateLimit(redisService))
			}

			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/logout", authHandler.Logout)
		})

		// Protected routes
		r.Route("/", func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware)

			// Add user-based rate limiting for authenticated users
			if redisService != nil {
				r.Use(customMiddleware.StrictRateLimit(redisService))
			}

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/profile", userHandler.GetProfile)
				r.Put("/profile", userHandler.UpdateProfile)
				r.Post("/photos", userHandler.UploadPhoto)
				r.Delete("/photos/{photoID}", userHandler.DeletePhoto)
				r.Put("/preferences", userHandler.UpdatePreferences)
				r.Get("/preferences", userHandler.GetPreferences)
			})

			// Match routes
			r.Route("/matches", func(r chi.Router) {
				r.Post("/swipe", matchHandler.Swipe)
				r.Get("/", matchHandler.GetMatches)
				r.Get("/potential", matchHandler.GetPotentialMatches)
				r.Delete("/{matchID}", matchHandler.UnMatch)
			})

			// Chat routes
			r.Route("/chats", func(r chi.Router) {
				r.Get("/", chatHandler.GetChats)
				r.Get("/{chatID}/messages", chatHandler.GetMessages)
				r.Post("/{chatID}/messages", chatHandler.SendMessage)
			})

			// Notification routes
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", notificationHandler.GetNotifications)
				r.Put("/{notificationID}/read", notificationHandler.MarkAsRead)
				r.Put("/read-all", notificationHandler.MarkAllAsRead)
				r.Get("/unread-count", notificationHandler.GetUnreadCount)
				r.Get("/preferences", notificationHandler.GetPreferences)
				r.Put("/preferences", notificationHandler.UpdatePreferences)
				r.Post("/devices", notificationHandler.RegisterDevice)
				r.Delete("/devices/{tokenID}", notificationHandler.UnregisterDevice)
				r.Post("/test", notificationHandler.TestNotification)
			})

			// Image routes
			r.Route("/images", func(r chi.Router) {
				r.Get("/", imageHandler.ListUserImages)
				r.Post("/upload", imageHandler.UploadImage)
				r.Post("/upload-base64", imageHandler.UploadImageBase64)
				r.Post("/presigned-upload", imageHandler.GeneratePresignedUploadURL)
				r.Get("/download/{imageKey}", imageHandler.DownloadImage)
				r.Delete("/{imageKey}", imageHandler.DeleteImage)
			})
		})

		// WebSocket endpoint
		r.Get("/ws", chatHandler.HandleWebSocket)
	})

	// Server configuration
	port := getEnvOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
