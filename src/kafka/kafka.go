package kafka

import (
	"app/infrastructure/postgres"
	"app/infrastructure/repository"
	kafka_hanlders "app/kafka/hanlders"
	usecase_user "app/usecase/user"
	"log"

	"github.com/segmentio/kafka-go"
)

func StartKafka() {

	db, err := postgres.Connect()

	if err != nil {
		log.Fatal(err)
	}

	repositoryUser := repository.NewUserPostgres(db)
	usecaseUser := usecase_user.NewService(repositoryUser)

	var topicParams []KafkaReadTopicsParams

	topicParams = append(topicParams, KafkaReadTopicsParams{
		Topic: "user",
		Handler: func(m kafka.Message) error {
			return kafka_hanlders.CreateUser(m, usecaseUser)
		},
	})

	startKafkaConnection(topicParams)
	readTopics()
}
