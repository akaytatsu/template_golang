package handlers

import (
	"app/api/middleware"
	"app/infrastructure/repository"
	usecase_user "app/usecase/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	}
	return false
}

func jsonResponse(c *gin.Context, httpStatus int, data any) {
	c.JSON(httpStatus, data)
}

func RoutersHandler(c *gin.Context, r *gin.Engine) {
	type Router struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}

	var routers []Router = make([]Router, 0)

	for _, route := range r.Routes() {
		routers = append(routers, Router{
			Method: route.Method,
			Path:   route.Path,
		})
	}

	if gin.Mode() == gin.DebugMode {
		c.JSON(200, routers)
	}
}

func SetAuthMiddleware(conn *gorm.DB, group *gin.RouterGroup) {
	usecaseUser := usecase_user.NewService(
		repository.NewUserPostgres(conn),
	)

	group.Use(middleware.AuthenticatedMiddleware(usecaseUser))
}
