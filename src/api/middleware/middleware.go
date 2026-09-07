package middleware

import (
	"app/entity"
	"net/http"
	"strings"

	usecase_user "app/usecase/user"

	"github.com/gin-gonic/gin"
)

// ContextUserKey é a chave sob a qual os middlewares guardam o usuário
// autenticado no contexto do Gin.
const ContextUserKey = "user"

// CurrentUser recupera o usuário autenticado colocado no contexto pelos
// middlewares. ok=false quando a rota não passou por um deles.
func CurrentUser(c *gin.Context) (entity.EntityUser, bool) {
	value, exists := c.Get(ContextUserKey)
	if !exists {
		return entity.EntityUser{}, false
	}

	user, ok := value.(entity.EntityUser)

	return user, ok
}

// bearerToken extrai o token de um header Authorization no formato
// "Bearer <token>". Retorna ok=false para header ausente, esquema diferente de
// Bearer ou token vazio.
func bearerToken(c *gin.Context) (string, bool) {
	scheme, token, found := strings.Cut(c.GetHeader("Authorization"), " ")
	if !found {
		return "", false
	}

	if !strings.EqualFold(strings.TrimSpace(scheme), "bearer") {
		return "", false
	}

	token = strings.TrimSpace(token)

	return token, token != ""
}

// authenticate resolve o usuário do header Authorization. Em qualquer falha
// responde 401 e aborta, com retorno nil.
func authenticate(c *gin.Context, usecase usecase_user.IUsecaseUser, message string) *entity.EntityUser {
	token, ok := bearerToken(c)
	if !ok {
		abortUnauthorized(c, message)
		return nil
	}

	user, err := usecase.GetUserByToken(token)
	if err != nil || user == nil {
		abortUnauthorized(c, "Unauthorized")
		return nil
	}

	return user
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": message})
}

func AuthenticatedMiddleware(usercase usecase_user.IUsecaseUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authenticate(c, usercase, "Unauthorized")
		if user == nil {
			return
		}

		c.Set(ContextUserKey, *user)
		c.Next()
	}
}

func AdminMiddleware(usercase usecase_user.IUsecaseUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authenticate(c, usercase, "Not Authenticated")
		if user == nil {
			return
		}

		if !user.IsAdmin {
			abortUnauthorized(c, "Unauthorized")
			return
		}

		c.Set(ContextUserKey, *user)
		c.Next()
	}
}
