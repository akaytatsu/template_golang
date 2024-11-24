package kafka

import (
	"app/config"
	"context"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	KafkaBootstrapServers string
	KafkaClientID         string
	KafkaGroupID          string
)

type KafkaReadTopicsParams struct {
	Topic   string
	Handler func(m *kafka.Message) error
}

var TopicParams []KafkaReadTopicsParams

func startKafkaConnection(topicParams []KafkaReadTopicsParams) {
	TopicParams = topicParams

	var topicConfigs []kafka.TopicSpecification
	KafkaBootstrapServers = config.EnvironmentVariables.KAFKA_BOOTSTRAP_SERVER
	KafkaClientID = config.EnvironmentVariables.KAFKA_CLIENT_ID
	KafkaGroupID = config.EnvironmentVariables.KAFKA_GROUP_ID

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": KafkaBootstrapServers})
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}
	defer adminClient.Close()

	for _, topicParam := range TopicParams {
		topicConfigs = append(topicConfigs, kafka.TopicSpecification{
			Topic:             topicParam.Topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	_, err = adminClient.CreateTopics(context.Background(), topicConfigs)
	if err != nil {
		log.Println("Error creating topic: ", err)
	}
}

func readTopics() {

	var topics []string
	for _, topicParam := range TopicParams {
		topics = append(topics, topicParam.Topic)
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": KafkaBootstrapServers,
		"group.id":          KafkaGroupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	defer consumer.Close()

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %v", err)
	}

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Println("Error while fetching message:", err)
			continue
		}
		var success bool = false

		for _, topicParam := range TopicParams {
			if topicParam.Topic == *msg.TopicPartition.Topic {

				if topicParam.Handler == nil {
					continue
				}

				// err = topicParam.Handler(msg)

				err := readMessageMiddlewareAPM(msg, topicParam.Handler)

				if err != nil {
					success = false
				} else {
					success = true
				}
			}
		}

		if success {
			_, err = consumer.CommitMessage(msg)
			if err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
		}
	}

}

func PublishMessage(topic string, message string) error {
	// producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": KafkaBootstrapServers})
	// if err != nil {
	// 	return err
	// }
	// defer producer.Close()

	// deliveryChan := make(chan kafka.Event, 1)

	// err = producer.Produce(&kafka.Message{
	// 	TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
	// 	Value:          []byte(message),
	// }, deliveryChan)
	// if err != nil {
	// 	return err
	// }

	// e := <-deliveryChan
	// m := e.(*kafka.Message)
	// if m.TopicPartition.Error != nil {
	// 	log.Printf("Delivery failed: %v", m.TopicPartition.Error)
	// } else {
	// 	log.Printf("Delivered message to %v", m.TopicPartition)
	// }
	// close(deliveryChan)

	// return nil

	return publishMessageAPM(topic, message)
}
