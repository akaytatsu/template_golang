package middleware

import (
	"net/http"
	"strings"

	usecase_user "app/usecase/user"

	"github.com/gin-gonic/gin"
)

const (
	bearerPrefix = "Bearer "
)

// extractBearerToken extracts the Bearer token from the Authorization header
func extractBearerToken(authHeader string) (string, bool) {
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", false
	}
	return strings.TrimPrefix(authHeader, bearerPrefix), true
}

func AuthenticatedMiddleware(usercase usecase_user.IUsecaseUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")

		token, valid := extractBearerToken(authHeader)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized: missing or invalid authorization header",
			})
			c.Abort()
			return
		}

		user, err := usercase.GetUserByToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized: invalid token",
			})
			c.Abort()
			return
		}

		// set user to context
		c.Set("user", *user)
		c.Next()
	}
}

func AdminMiddleware(usercase usecase_user.IUsecaseUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")

		token, valid := extractBearerToken(authHeader)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized: missing or invalid authorization header",
			})
			c.Abort()
			return
		}

		user, err := usercase.GetUserByToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized: invalid token",
			})
			c.Abort()
			return
		}

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Forbidden: admin access required",
			})
			c.Abort()
			return
		}

		// set user to context
		c.Set("user", *user)
		c.Next()
	}
}
