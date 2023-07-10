package handlers

import (
	usecase_user "app/usecase/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(c *gin.Context, usecaseUser usecase_user.IUsecaseUser) {

	var loginData LoginData

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	user, err := usecaseUser.LoginUser(loginData.Email, loginData.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	token, refreshToken, err := user.JWTTokenGenerator()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "refreshToken": refreshToken})

}
