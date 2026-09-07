package handlers_test

import (
	"app/api/handlers"
	"app/api/middleware"
	"app/entity"
	"app/mocks"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequestWithContext(context.Background(), method, path, nil)
	} else {
		req = httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// Regressão: handleError devolvia 500 para tudo — payload inválido, credencial
// errada e registro inexistente ficavam indistinguíveis.
func TestLoginHandler_StatusCodes(t *testing.T) {
	t.Run("payload malformado devolve 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		usecase := mocks.NewMockIUsecaseUser(ctrl)

		r := gin.New()
		r.POST("/login", handlers.NewUserHandler(usecase).LoginHandler)

		w := do(r, http.MethodPost, "/login", `{"email": `)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("senha errada devolve 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		usecase := mocks.NewMockIUsecaseUser(ctrl)
		usecase.EXPECT().LoginUser("a@b.com", "errada").
			Return(nil, bcryptMismatch())

		r := gin.New()
		r.POST("/login", handlers.NewUserHandler(usecase).LoginHandler)

		w := do(r, http.MethodPost, "/login", `{"email":"a@b.com","password":"errada"}`)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usuario inexistente devolve 404", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		usecase := mocks.NewMockIUsecaseUser(ctrl)
		usecase.EXPECT().LoginUser("nao@existe.com", "x").
			Return(nil, gorm.ErrRecordNotFound)

		r := gin.New()
		r.POST("/login", handlers.NewUserHandler(usecase).LoginHandler)

		w := do(r, http.MethodPost, "/login", `{"email":"nao@existe.com","password":"x"}`)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("erro inesperado devolve 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		usecase := mocks.NewMockIUsecaseUser(ctrl)
		usecase.EXPECT().LoginUser("a@b.com", "x").Return(nil, errors.New("boom"))

		r := gin.New()
		r.POST("/login", handlers.NewUserHandler(usecase).LoginHandler)

		w := do(r, http.MethodPost, "/login", `{"email":"a@b.com","password":"x"}`)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func bcryptMismatch() error {
	// bcrypt.ErrMismatchedHashAndPassword, obtido via a própria entity.
	u := entity.EntityUser{Password: "$2a$10$invalidhashinvalidhashinvalidhashinvalidhashinvalidhashinv"}

	return u.ValidatePassword("qualquer")
}

// Regressão: cada handler fazia strconv.Atoi descartando o erro, então
// /api/user/abc virava o id 0 silenciosamente.
func TestGetUserHandler_InvalidPathID(t *testing.T) {
	for _, id := range []string{"abc", "0", "-1", "1.5"} {
		t.Run(id, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Id inválido: o usecase não deve ser chamado.
			usecase := mocks.NewMockIUsecaseUser(ctrl)

			r := gin.New()
			r.GET("/user/:id", handlers.NewUserHandler(usecase).GetUserHandler)

			w := do(r, http.MethodGet, "/user/"+id, "")

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGetUserHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUser(7).Return(&entity.EntityUser{ID: 7, Email: "a@b.com"}, nil)

	r := gin.New()
	r.GET("/user/:id", handlers.NewUserHandler(usecase).GetUserHandler)

	w := do(r, http.MethodGet, "/user/7", "")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "a@b.com")
}

// Regressão: UpdateUserHandler lia c.GetInt("id") (contexto, não path param),
// então o update sempre mirava o registro 0.
func TestUpdateUserHandler_UsesPathID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var received *entity.EntityUser

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().Update(gomock.Any()).
		DoAndReturn(func(u *entity.EntityUser) error {
			received = u
			return nil
		})

	r := gin.New()
	r.PUT("/user/:id", handlers.NewUserHandler(usecase).UpdateUserHandler)

	w := do(r, http.MethodPut, "/user/42", `{"name":"Novo","email":"novo@example.com"}`)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, received)
	assert.Equal(t, 42, received.ID, "o id tem de vir do path param, não ser 0")
}

// Regressão: DeleteUserHandler exigia o usuário no corpo de um DELETE.
func TestDeleteUserHandler_UsesPathID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stored := &entity.EntityUser{ID: 9, Email: "del@example.com"}

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUser(9).Return(stored, nil)
	usecase.EXPECT().Delete(stored).Return(nil)

	r := gin.New()
	r.DELETE("/user/:id", handlers.NewUserHandler(usecase).DeleteUserHandler)

	w := do(r, http.MethodDelete, "/user/9", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

// Regressão: GetMeHandler passava o header inteiro ("Bearer <token>") para
// ValidateToken, que falhava sempre. Agora lê o usuário do contexto.
func TestGetMeHandler_ReadsUserFromContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)

	r := gin.New()
	r.GET("/me", func(c *gin.Context) {
		c.Set(middleware.ContextUserKey, entity.EntityUser{ID: 3, Email: "me@example.com"})
	}, handlers.NewUserHandler(usecase).GetMeHandler)

	w := do(r, http.MethodGet, "/me", "")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "me@example.com")
}

func TestGetMeHandler_WithoutMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockIUsecaseUser(ctrl)

	r := gin.New()
	r.GET("/me", handlers.NewUserHandler(usecase).GetMeHandler)

	w := do(r, http.MethodGet, "/me", "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// A listagem agora pagina de verdade e devolve PaginationResponse.
func TestGetUsersHandler_Pagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var received entity.EntityUserFilters

	usecase := mocks.NewMockIUsecaseUser(ctrl)
	usecase.EXPECT().GetUsers(gomock.Any()).
		DoAndReturn(func(f entity.EntityUserFilters) ([]entity.EntityUser, int64, error) {
			received = f
			return []entity.EntityUser{{ID: 1}, {ID: 2}}, int64(25), nil
		})

	r := gin.New()
	r.GET("/list", handlers.NewUserHandler(usecase).GetUsersHandler)

	w := do(r, http.MethodGet, "/list?page=2&page_size=10&search=ana", "")
	require.Equal(t, http.StatusOK, w.Code)

	assert.Equal(t, 2, received.Page)
	assert.Equal(t, 10, received.PageSize)
	assert.Equal(t, "ana", received.Search)

	var resp handlers.PaginationResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, 3, resp.TotalPages, "ceil(25/10) = 3")
	assert.Equal(t, 25, resp.TotalRegisters)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
}

func TestGetPaginationParams_Limits(t *testing.T) {
	cases := map[string]struct {
		query            string
		wantPage, wantSz int
	}{
		"padrao":                 {"", 0, 10},
		"page_size acima do max": {"?page_size=5000", 0, 100},
		"page_size zero":         {"?page_size=0", 0, 10},
		"page_size negativo":     {"?page_size=-5", 0, 10},
		"page negativo":          {"?page=-3", 0, 10},
		"valores validos":        {"?page=4&page_size=25", 4, 25},
		"page_size invalido":     {"?page_size=abc", 0, 10},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var page, size int

			r := gin.New()
			r.GET("/x", func(c *gin.Context) {
				page, size = handlers.GetPaginationParams(c)
				c.Status(http.StatusOK)
			})

			do(r, http.MethodGet, "/x"+tc.query, "")

			assert.Equal(t, tc.wantPage, page)
			assert.Equal(t, tc.wantSz, size)
		})
	}
}

func TestGetTotalPages(t *testing.T) {
	assert.Equal(t, 3, handlers.GetTotalPages(25, 10))
	assert.Equal(t, 2, handlers.GetTotalPages(20, 10))
	assert.Equal(t, 0, handlers.GetTotalPages(0, 10))
	assert.Equal(t, 1, handlers.GetTotalPages(1, 10))
	// pageSize zero não deve causar divisão por zero nem +Inf.
	assert.Equal(t, 0, handlers.GetTotalPages(25, 0))
}
