package kafka_hanlders

import (
	"app/entity"
	usecase_user "app/usecase/user"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

func CreateUser(m kafka.Message, usecaseUser usecase_user.IUsecaseUser) error {

	var entityUser entity.EntityUser

	err := json.Unmarshal(m.Value, &entityUser)

	if err != nil {
		return err
	}

	err = usecaseUser.Create(&entityUser)

	return err
}
