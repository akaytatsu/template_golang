package api

import (
	"app/api/handlers"
	"app/config"
	"app/infrastructure/postgres"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "app/docs"

	custom_logger "app/pkg/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupDatabase() *gorm.DB {
	conn := postgres.Connect()
	return conn
}

func setupRouter(conn *gorm.DB) *gin.Engine {
	gin.SetMode(config.EnvironmentVariables.GinMode)

	r := gin.New()

	corsConfig := cors.DefaultConfig()
	// Configure allowed origins based on environment
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		corsConfig.AllowOrigins = []string{allowedOrigins}
	} else {
		// Default to localhost for development
		corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	}
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("authorization", "content-type")

	r.Use(cors.New(corsConfig))

	// Configurar middleware de logging baseado no nível de log
	if config.EnvironmentVariables.GinMode == "debug" || custom_logger.ShouldLogLevel(config.EnvironmentVariables.LogLevel, "INFO") {
		r.Use(gin.Logger())
	}

	r.Use(gin.Recovery())

	handlers.MountSamplesHandlers(r)
	handlers.MountUsersHandlers(r, conn)

	return r
}

func SetupRouters() *gin.Engine {
	conn := setupDatabase()
	return setupRouter(conn)
}

// StartWebServer starts the web server with graceful shutdown support
func StartWebServer() {
	r := SetupRouters()

	// Use the base URL from the environment or default to localhost
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	url := ginSwagger.URL(baseURL + "/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// se for release, reduz o log
	if config.EnvironmentVariables.ISRELEASE {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create HTTP server with timeout
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
