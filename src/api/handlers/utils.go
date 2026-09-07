package handlers

import (
	"app/api/middleware"
	"app/entity"
	"app/infrastructure/repository"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	usecase_user "app/usecase/user"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

type PaginationResponse struct {
	TotalPages     int `json:"total_pages"`
	Page           int `json:"page"`
	PageSize       int `json:"page_size"`
	TotalRegisters int `json:"total_registers"`
	Registers      any `json:"registers"`
}

// httpStatusFor traduz um erro de domínio no status HTTP adequado.
// Antes tudo virava 500, o que escondia erro de payload, credencial inválida e
// registro inexistente atrás do mesmo código.
func httpStatusFor(err error) int {
	var validationErrors validator.ValidationErrors

	switch {
	case errors.As(err, &validationErrors):
		return http.StatusBadRequest
	case errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return http.StatusUnauthorized
	case errors.Is(err, entity.ErrInvalidToken), errors.Is(err, entity.ErrInvalidClaims):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func handleError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	status := httpStatusFor(err)
	body := gin.H{"error": err.Error()}

	// Em erro de validação, devolve também o detalhe por campo.
	if fields := entity.GetStructError(err); len(fields) > 0 {
		body["fields"] = fields
	}

	c.AbortWithStatusJSON(status, body)

	return true
}

// handleBindError responde 400 para payload malformado.
func handleBindError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	return true
}

func jsonResponse(c *gin.Context, httpStatus int, data any) {
	c.JSON(httpStatus, data)
}

// pathID lê o path param `id`. Antes cada handler fazia strconv.Atoi
// descartando o erro, então um id não numérico virava 0 silenciosamente.
func pathID(c *gin.Context) (int, error) {
	raw := c.Param("id")

	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("id inválido: %q", raw)
	}

	if id <= 0 {
		return 0, fmt.Errorf("id inválido: %q", raw)
	}

	return id, nil
}

func RoutersHandler(c *gin.Context, r *gin.Engine) {
	type Router struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}

	routers := make([]Router, 0)

	for _, route := range r.Routes() {
		routers = append(routers, Router{
			Method: route.Method,
			Path:   route.Path,
		})
	}

	if gin.Mode() == gin.DebugMode {
		c.JSON(http.StatusOK, routers)
	}
}

func SetAuthMiddleware(conn *gorm.DB, group *gin.RouterGroup) {
	usecaseUser := usecase_user.NewService(
		repository.NewUserPostgres(conn),
	)

	group.Use(middleware.AuthenticatedMiddleware(usecaseUser))
}

func SetAdminMiddleware(conn *gorm.DB, group *gin.RouterGroup) {
	usecaseUser := usecase_user.NewService(
		repository.NewUserPostgres(conn),
	)

	group.Use(middleware.AdminMiddleware(usecaseUser))
}

// GetPaginationParams lê `page` e `page_size` da query string, aplicando os
// limites padrão.
func GetPaginationParams(c *gin.Context) (page, pageSize int) {
	page = 0
	pageSize = defaultPageSize

	if c.Query("page") != "" {
		page, _ = strconv.Atoi(c.Query("page"))
		if page < 0 {
			page = 0
		}
	}

	if c.Query("page_size") != "" {
		pageSize, _ = strconv.Atoi(c.Query("page_size"))
		if pageSize < 1 {
			pageSize = defaultPageSize
		}

		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
	}

	return page, pageSize
}

// GetOrderAndSortByParams lê `order_by` e `sort_order` da query string,
// caindo nos valores padrão quando ausentes.
func GetOrderAndSortByParams(c *gin.Context, defaultOrder string, defaultSort string) (orderBy, sortOrder string) {
	orderBy = c.Query("order_by")
	sortOrder = c.Query("sort_order")

	if orderBy == "" {
		orderBy = defaultOrder
	}

	if sortOrder == "" {
		sortOrder = defaultSort
	}

	return orderBy, sortOrder
}

// GetOrderByParams é o GetOrderAndSortByParams com ordenação ascendente.
func GetOrderByParams(c *gin.Context, defaultValue string) (orderBy, sortOrder string) {
	return GetOrderAndSortByParams(c, defaultValue, "asc")
}

// GetTotalPages calcula o número de páginas para um total de registros.
func GetTotalPages(totalRegisters int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}

	return int(math.Ceil(float64(totalRegisters) / float64(pageSize)))
}
