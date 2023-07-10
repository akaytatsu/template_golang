package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthenticatedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get bearer token from header
		token := c.Request.Header.Get("Authorization")

		// check if token is valid
		if token == "Bearer x" {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
		}
	}
}
