package kafka

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

var (
	KafkaBootstrapServers string
	KafkaClientID         string
	KafkaGroupID          string
)

type KafkaReadTopicsParams struct {
	Topic   string
	Handler func(m kafka.Message) error
}

var TopicParams []KafkaReadTopicsParams

func startKafkaConnection(topicParams []KafkaReadTopicsParams) {
	TopicParams = topicParams

	var topicConfigs []kafka.TopicConfig
	KafkaBootstrapServers = os.Getenv("KAFKA_BOOTSTRAP_SERVER")
	KafkaClientID = os.Getenv("KAFKA_CLIENT_ID")
	KafkaGroupID = os.Getenv("KAFKA_GROUP_ID")

	var controllerConn *kafka.Conn

	controllerConn, err := kafka.Dial("tcp", strings.Split(KafkaBootstrapServers, ",")[0])
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	defer controllerConn.Close()

	for _, topicParam := range TopicParams {

		topicConfigs = []kafka.TopicConfig{
			{
				Topic:             topicParam.Topic,
				NumPartitions:     1,
				ReplicationFactor: -1,
			},
		}
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Println("Error creating topic: ", err)
	}
}

func readTopics() {
	l := logrus.New()

	var topics []string
	for _, topicParam := range TopicParams {
		topics = append(topics, topicParam.Topic)
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(KafkaBootstrapServers, ","),
		GroupID:     KafkaGroupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
		// MaxWait:     1 * time.Second,
		GroupTopics: topics,
		// Logger:      kafka.LoggerFunc(l.Infof),
		ErrorLogger: kafka.LoggerFunc(l.Errorf),
	})

	defer r.Close()

	ctx := context.Background()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			log.Println("Error while fetching message:", err)
			continue
		}
		var success bool = false

		for _, topicParam := range TopicParams {
			if topicParam.Topic == m.Topic {

				if topicParam.Handler == nil {
					continue
				}

				err = topicParam.Handler(m)

				if err != nil {
					success = false
				} else {
					success = true
				}
			}
		}

		if success {
			r.CommitMessages(ctx, m)
		}
	}

}

func PublishMessage(topic string, message string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: strings.Split(KafkaBootstrapServers, ","),
		Topic:   topic,
	})

	defer writer.Close()

	messageToSend := kafka.Message{
		Value: []byte(message),
	}

	return writer.WriteMessages(context.Background(), messageToSend)
}
