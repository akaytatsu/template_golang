package handlers_test

import (
	"app/api/handlers"
	"app/entity"
	"app/mocks"
	"app/pkg/testing_utils"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
)

func TestHandlers_LoginHandle(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecaseUser := mocks.NewMockIUsecaseUser(ctrl)

	mockUsecaseUser.EXPECT().LoginUser("mailer@mailer.com", gomock.Any()).Return(nil, errors.New("error"))
	mockUsecaseUser.EXPECT().LoginUser("mailer33@mailer.com", "password").Return(&entity.EntityUser{
		Email: "mailer33@mailer.com",
	}, nil)

	convey.Convey("Invalid JSON Sender", t, func() {

		testRequest := testing_utils.TestRequest{
			Method: "POST",
			Url:    "/login",
			T:      t,
		}
		testRequest.SetTextBody("teste")

		testRequest.SetHandle(func(gin *gin.Context) {
			handlers.LoginHandler(gin, mockUsecaseUser)

		})

		testRequest.Execute()
		convey.So(testRequest.Response.Result().StatusCode, convey.ShouldEqual, http.StatusBadRequest)

	})

	convey.Convey("Mail wrong", t, func() {

		testRequest := testing_utils.TestRequest{
			Method: "POST",
			Url:    "/login",
			T:      t,
		}
		testRequest.SetJsonBody(map[string]any{
			"email":    "mailer@mailer.com",
			"password": "password",
		})

		testRequest.SetHandle(func(gin *gin.Context) {
			handlers.LoginHandler(gin, mockUsecaseUser)
		})

		testRequest.Execute()

		convey.So(testRequest.Response.Result().StatusCode, convey.ShouldEqual, http.StatusBadRequest)

	})

	convey.Convey("Success generate Token", t, func() {

		testRequest := testing_utils.TestRequest{
			Method: "POST",
			Url:    "/login",
			T:      t,
		}
		testRequest.SetJsonBody(map[string]any{
			"email":    "mailer33@mailer.com",
			"password": "password",
		})

		testRequest.SetHandle(func(gin *gin.Context) {
			handlers.LoginHandler(gin, mockUsecaseUser)
		})

		testRequest.Execute()

		convey.So(testRequest.Response.Result().StatusCode, convey.ShouldEqual, http.StatusOK)

	})

}
