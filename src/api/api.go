package api

import (
	"log"

	"app/api/handlers"
	"app/api/middleware"
	"app/config"
	"app/infrastructure/postgres"
	"app/infrastructure/repository"
	usecase_user "app/usecase/user"

	"github.com/gin-gonic/gin"
)

func SetupRouters() *gin.Engine {

	db, err := postgres.Connect()

	if err != nil {
		log.Fatal(err)
	}

	repositoryUser := repository.NewUserPostgres(db)
	usecaseUser := usecase_user.NewService(repositoryUser)

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/", handlers.HomeHandler)

	r.GET("/xml", handlers.XmlHandler)
	r.GET("/text", handlers.TextHandler)
	r.GET("/yaml", handlers.YamlHandler)
	r.GET("/protobuf", handlers.ProtobufHandler)
	r.GET("/sse", handlers.ServerSideEventsHandler)

	// Login do usuario
	r.POST("/login", func(gin *gin.Context) {
		handlers.LoginHandler(gin, usecaseUser)
	})

	authorized := r.Group("/api")

	authorized.Use(middleware.AuthenticatedMiddleware())

	authorized.GET("/secret", handlers.SecretHandler)

	return r
}

func StartWebServer() {
	config.ReadEnvironmentVars()

	r := SetupRouters()

	// Bind to a port and pass our router in
	log.Fatal(r.Run())
}
