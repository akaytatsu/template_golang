package handlers

import (
	"app/api/middleware"
	"app/entity"
	"app/infrastructure/repository"
	"net/http"

	usecase_user "app/usecase/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserPasswordData struct {
	Email           string `json:"email"`
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

type UserHandlers struct {
	UsecaseUser usecase_user.IUsecaseUser
}

func NewUserHandler(usecaseUser usecase_user.IUsecaseUser) *UserHandlers {
	return &UserHandlers{UsecaseUser: usecaseUser}
}

// @Summary Login
// @Description Login
// @Tags User
// @Accept  json
// @Produce  json
// @Param email body string true "Email"
// @Param password body string true "Password"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/login [post]
func (h UserHandlers) LoginHandler(c *gin.Context) {
	var loginData LoginData

	if handleBindError(c, c.ShouldBindJSON(&loginData)) {
		return
	}

	user, err := h.UsecaseUser.LoginUser(loginData.Email, loginData.Password)

	if exception := handleError(c, err); exception {
		return
	}

	token, refreshToken, err := user.JWTTokenGenerator()

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"token": token, "refreshToken": refreshToken})
}

// @Summary Get me
// @Description Get me
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/me [get]
func (h UserHandlers) GetMeHandler(c *gin.Context) {
	// O usuário já foi resolvido pelo AuthenticatedMiddleware. Antes este
	// handler repassava o header inteiro ("Bearer <token>") para ValidateToken,
	// que falhava sempre por token malformado.
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	jsonResponse(c, http.StatusOK, user)
}

// @Summary Create user
// @Description Create user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param entity.EntityUser body entity.EntityUser true "User"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/create [post]
func (h UserHandlers) CreateUserHandler(c *gin.Context) {
	var entityUser entity.EntityUser

	if handleBindError(c, c.ShouldBindJSON(&entityUser)) {
		return
	}

	err := h.UsecaseUser.Create(&entityUser)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "User created successfully"})
}

// @Summary Update user
// @Description Update user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Param entity.EntityUser body entity.EntityUser true "User"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/{id} [put]
func (h UserHandlers) UpdateUserHandler(c *gin.Context) {
	// c.GetInt("id") lia o contexto do Gin, onde "id" nunca é colocado — o id
	// vem do path param, então o update sempre mirava o registro 0.
	id, err := pathID(c)
	if err != nil {
		handleBindError(c, err)
		return
	}

	var entityUser entity.EntityUser

	if handleBindError(c, c.ShouldBindJSON(&entityUser)) {
		return
	}

	// Depois do bind: o id da rota é a fonte da verdade, não o corpo.
	entityUser.ID = id

	err = h.UsecaseUser.Update(&entityUser)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "User updated successfully"})
}

// @Summary Delete user
// @Description Delete user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/{id} [delete]
func (h UserHandlers) DeleteUserHandler(c *gin.Context) {
	// A rota é DELETE /api/user/:id, mas o handler exigia o usuário no corpo da
	// requisição — e o repositório resolve o registro pelo e-mail. Buscar pelo
	// id da rota e então remover.
	id, err := pathID(c)
	if err != nil {
		handleBindError(c, err)
		return
	}

	user, err := h.UsecaseUser.GetUser(id)

	if exception := handleError(c, err); exception {
		return
	}

	if err := h.UsecaseUser.Delete(user); handleError(c, err) {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// @Summary Update password
// @Description Update password
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Param entity.EntityUser body entity.EntityUser true "User"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/password/{id} [put]
func (h UserHandlers) UpdatePasswordHandler(c *gin.Context) {
	var updatePasswordData UpdateUserPasswordData

	if handleBindError(c, c.ShouldBindJSON(&updatePasswordData)) {
		return
	}

	id, err := pathID(c)
	if err != nil {
		handleBindError(c, err)
		return
	}

	err = h.UsecaseUser.UpdatePassword(id, updatePasswordData.OldPassword, updatePasswordData.NewPassword, updatePasswordData.ConfirmPassword)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// @Summary Get users
// @Description Get users
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param search query string false "Search"
// @Param active query string false "Active"
// @Param page query int false "Page (0-indexed)"
// @Param page_size query int false "Page size (max 100)"
// @Success 200 {object} PaginationResponse "success"
// @Router /api/user/list [get]
func (h UserHandlers) GetUsersHandler(c *gin.Context) {
	page, pageSize := GetPaginationParams(c)

	filters := entity.EntityUserFilters{
		Search:   c.Query("search"),
		Active:   c.Query("active"),
		Page:     page,
		PageSize: pageSize,
	}

	users, total, err := h.UsecaseUser.GetUsers(filters)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, PaginationResponse{
		TotalPages:     GetTotalPages(total, pageSize),
		Page:           page,
		PageSize:       pageSize,
		TotalRegisters: int(total),
		Registers:      users,
	})
}

// @Summary Get user
// @Description Get user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/{id} [get]
func (h UserHandlers) GetUserHandler(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		handleBindError(c, err)
		return
	}

	user, err := h.UsecaseUser.GetUser(id)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, user)
}

func MountUsersHandlers(gin *gin.Engine, conn *gorm.DB) {
	userHandlers := NewUserHandler(
		usecase_user.NewService(
			repository.NewUserPostgres(conn),
		),
	)

	gin.GET("/", HomeHandler)
	gin.POST("/api/login", userHandlers.LoginHandler)

	gin.POST("/login", userHandlers.LoginHandler)

	// user
	group := gin.Group("/api/user")
	SetAuthMiddleware(conn, group)

	group.GET("/me", userHandlers.GetMeHandler)
	group.POST("/create", userHandlers.CreateUserHandler)
	group.PUT("/:id", userHandlers.UpdateUserHandler)
	group.DELETE("/:id", userHandlers.DeleteUserHandler)
	group.PUT("/password/:id", userHandlers.UpdatePasswordHandler)
	group.GET("/list", userHandlers.GetUsersHandler)
	group.GET("/:id", userHandlers.GetUserHandler)
}
