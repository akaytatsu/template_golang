package kafka

import (
	"app/infrastructure/repository"
	"context"

	kafka_handlers "app/kafka/handlers"
	usecase_user "app/usecase/user"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"gorm.io/gorm"
)

// StartKafka configura os tópicos e consome mensagens até o context ser
// cancelado. Bloqueia — chame numa goroutine.
func StartKafka(ctx context.Context, db *gorm.DB) {
	repositoryUser := repository.NewUserPostgres(db)
	usecaseUser := usecase_user.NewService(repositoryUser)

	topicParams := []KafkaReadTopicsParams{
		{
			Topic: "user",
			Handler: func(msg *kafka.Message) error {
				return kafka_handlers.CreateUser(*msg, usecaseUser)
			},
		},
	}

	kafkaSetup(topicParams)

	// Só inicia leitura de tópicos se houver tópicos configurados e conexão estabelecida
	if len(topicParams) > 0 && KafkaBootstrapServers != "" {
		readTopics(ctx, topicParams)
	}
}
