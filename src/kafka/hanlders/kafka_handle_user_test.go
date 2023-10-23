package kafka_hanlders_test

import (
	kafka_hanlders "app/kafka/hanlders"
	"app/mocks"
	"app/pkg/utils"
	usecase_user "app/usecase/user"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/smartystreets/goconvey/convey"
)

func TestKafkaHandleUser_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().CreateUser(gomock.Any()).Return(nil)

	convey.Convey("Test KafkaHandleUser CreateUser failed", t, func() {

		usecaseUser := usecase_user.NewService(mockUserRepo)

		message := kafka.Message{
			Value: utils.GenericMapToJson(map[string]any{}),
		}

		err := kafka_hanlders.CreateUser(message, usecaseUser)

		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("Test KafkaHandleUser CreateUser success", t, func() {

		usecaseUser := usecase_user.NewService(mockUserRepo)

		message := kafka.Message{
			Value: utils.GenericMapToJson(map[string]any{
				"name":     "name",
				"email":    "test@test.com",
				"password": "test587444",
			}),
		}

		err := kafka_hanlders.CreateUser(message, usecaseUser)

		convey.So(err, convey.ShouldBeNil)
	})
}
