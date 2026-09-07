package middleware_test

import (
	"app/api/middleware"
	"app/entity"
	"app/mocks"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// run monta uma rota protegida pelo middleware e devolve a resposta e se o
// handler final chegou a ser executado.
func run(t *testing.T, mw gin.HandlerFunc, authorization string) (*httptest.ResponseRecorder, bool) {
	t.Helper()

	reached := false

	r := gin.New()
	r.GET("/protegida", mw, func(c *gin.Context) {
		reached = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protegida", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w, reached
}

// Regressão: os middlewares chamavam c.Abort() sem `return` e caíam em
// strings.Split(header, " ")[1], estourando index out of range.
func TestAuthenticatedMiddleware_MalformedHeaderDoesNotPanic(t *testing.T) {
	headers := map[string]string{
		"ausente":            "",
		"sem espaco":         "xyz",
		"bearer sem token":   "Bearer",
		"bearer token vazio": "Bearer ",
		"esquema errado":     "Basic dXNlcjpwYXNz",
	}

	for name, header := range headers {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Header inválido: o usecase não deve nem ser consultado.
			usecase := mocks.NewMockIUsecaseUser(ctrl)

			assert.NotPanics(t, func() {
				w, reached := run(t, middleware.AuthenticatedMiddleware(usecase), header)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
				assert.False(t, reached, "handler protegido não deve executar")
			})
		})
	}
}

func TestAuthenticatedMiddleware_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUserByToken("token-invalido").Return(nil, errors.New("invalid token"))

	w, reached := run(t, middleware.AuthenticatedMiddleware(usecase), "Bearer token-invalido")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, reached)
}

// Regressão: com o repositório engolindo o erro de not-found, o usecase podia
// devolver (nil, nil) e o middleware fazia *user num ponteiro nil.
func TestAuthenticatedMiddleware_NilUserDoesNotPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUserByToken("token").Return(nil, nil)

	assert.NotPanics(t, func() {
		w, reached := run(t, middleware.AuthenticatedMiddleware(usecase), "Bearer token")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, reached)
	})
}

func TestAuthenticatedMiddleware_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := &entity.EntityUser{ID: 7, Email: "user@example.com"}

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUserByToken("token-ok").Return(expected, nil)

	var fromContext entity.EntityUser
	var ok bool

	r := gin.New()
	r.GET("/protegida", middleware.AuthenticatedMiddleware(usecase), func(c *gin.Context) {
		fromContext, ok = middleware.CurrentUser(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protegida", nil)
	req.Header.Set("Authorization", "Bearer token-ok")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, ok, "CurrentUser deve encontrar o usuário no contexto")
	assert.Equal(t, 7, fromContext.ID)
}

func TestAdminMiddleware_RejectsNonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUserByToken("token").Return(&entity.EntityUser{ID: 1, IsAdmin: false}, nil)

	w, reached := run(t, middleware.AdminMiddleware(usecase), "Bearer token")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, reached)
}

func TestAdminMiddleware_AllowsAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUserByToken("token").Return(&entity.EntityUser{ID: 1, IsAdmin: true}, nil)

	w, reached := run(t, middleware.AdminMiddleware(usecase), "Bearer token")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, reached)
}

func TestAdminMiddleware_MalformedHeaderDoesNotPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)

	assert.NotPanics(t, func() {
		w, reached := run(t, middleware.AdminMiddleware(usecase), "xyz")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, reached)
	})
}

func TestCurrentUser_WithoutMiddleware(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	_, ok := middleware.CurrentUser(c)

	assert.False(t, ok)
}
