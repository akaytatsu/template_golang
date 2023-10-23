package api

import (
	"log"

	"app/api/handlers"
	"app/config"
	"app/infrastructure/postgres"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupDatabase() *gorm.DB {
	conn := postgres.Connect()
	return conn
}

func setupRouter(conn *gorm.DB) *gin.Engine {
	r := gin.New()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AddAllowHeaders("authorization")

	r.Use(cors.New(config))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	handlers.MountSamplesHandlers(r)
	handlers.MountUsersHandlers(r, conn)

	return r
}

func SetupRouters() *gin.Engine {
	conn := setupDatabase()
	return setupRouter(conn)
}

func StartWebServer() {
	config.ReadEnvironmentVars()

	r := SetupRouters()

	// Bind to a port and pass our router in
	log.Fatal(r.Run())
}
