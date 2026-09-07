package api

import (
	"app/api/handlers"
	"app/config"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "app/docs"

	custom_logger "app/pkg/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	defaultPort = "8080"

	// readHeaderTimeout limita quanto tempo um cliente pode levar para enviar
	// os headers, mitigando Slowloris (gosec G112).
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second

	// shutdownTimeout é quanto esperamos as requisições em voo terminarem.
	shutdownTimeout = 15 * time.Second

	healthcheckTimeout = 5 * time.Second
)

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}

	return defaultPort
}

func setupRouter(conn *gorm.DB) *gin.Engine {
	gin.SetMode(config.EnvironmentVariables.GinMode)

	r := gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("authorization")

	r.Use(cors.New(corsConfig))

	// Configurar middleware de logging baseado no nível de log
	if config.EnvironmentVariables.GinMode == "debug" || custom_logger.ShouldLogLevel(config.EnvironmentVariables.LogLevel, "INFO") {
		r.Use(gin.Logger())
	}

	r.Use(gin.Recovery())

	handlers.MountSamplesHandlers(r)
	handlers.MountUsersHandlers(r, conn)

	// Health check endpoint
	r.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ginSwagger.URL(fmt.Sprintf("http://localhost:%s/swagger/doc.json", port())) // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// SetupRouters monta o router sobre uma conexão já estabelecida.
func SetupRouters(conn *gorm.DB) *gin.Engine {
	return setupRouter(conn)
}

// StartWebServer sobe o servidor HTTP e bloqueia até o context ser cancelado,
// quando faz o shutdown ordenado.
func StartWebServer(ctx context.Context, conn *gorm.DB) error {
	r := SetupRouters(conn)

	// se for release, reduz o log
	if config.EnvironmentVariables.ISRELEASE {
		gin.SetMode(gin.ReleaseMode)
	}

	srv := &http.Server{
		Addr:              net.JoinHostPort("", port()),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("servidor http escutando em %s", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Println("sinal recebido, encerrando o servidor http...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	}
}

// Healthcheck bate no endpoint /health da instância local. Usado pela flag
// -healthcheck do binário, que é o que o HEALTHCHECK do Dockerfile invoca.
func Healthcheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()

	url := fmt.Sprintf("http://localhost:%s/health", port())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status inesperado: %d", resp.StatusCode)
	}

	return nil
}
