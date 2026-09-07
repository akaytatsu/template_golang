package kafka

import (
	"app/config"
	"context"
	"errors"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	KafkaBootstrapServers string
	KafkaClientID         string
	KafkaGroupID          string
)

// producer é criado uma vez no kafkaSetup e reusado por PublishMessage. Criar
// um producer por mensagem custa handshake + fetch de metadata a cada envio.
var producer *kafka.Producer

type KafkaReadTopicsParams struct {
	Topic   string
	Handler func(m *kafka.Message) error
}

// Configurações padrão para os tópicos
const (
	NumPartitions     = 3
	ReplicationFactor = -1

	// readPollTimeout é o intervalo em que o consumer devolve o controle para
	// que o cancelamento do context seja observado.
	readPollTimeout = 500 * time.Millisecond
)

func kafkaSetup(topicParams []KafkaReadTopicsParams) {
	KafkaBootstrapServers = config.EnvironmentVariables.KAFKA_BOOTSTRAP_SERVER
	KafkaClientID = config.EnvironmentVariables.KAFKA_CLIENT_ID
	KafkaGroupID = config.EnvironmentVariables.KAFKA_GROUP_ID

	if KafkaBootstrapServers == "" {
		log.Println("KAFKA_BOOTSTRAP_SERVER não configurado, Kafka desabilitado")
		return
	}

	ensureTopics(KafkaBootstrapServers, topicParams)

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": KafkaBootstrapServers,
		"client.id":         KafkaClientID,
	})
	if err != nil {
		log.Printf("Erro ao criar o producer: %v", err)
		return
	}
	producer = p

	log.Println("Kafka configurado com sucesso")
}

// Close libera o producer compartilhado. Deve ser chamado no shutdown.
func Close() {
	if producer != nil {
		producer.Close()
		producer = nil
	}
}

func ensureTopics(broker string, topicParams []KafkaReadTopicsParams) {
	// Timeout administration (exemplo: 30 segundos)
	adminTimeout := 30 * time.Second

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		// Sem o return, o defer abaixo rodaria sobre um ponteiro nil.
		log.Printf("Failed to create AdminClient: %s\n", err)
		return
	}
	defer adminClient.Close()

	// Criação dos tópicos
	topicSpecifications := make([]kafka.TopicSpecification, 0, len(topicParams))
	for _, param := range topicParams {
		topicSpecifications = append(topicSpecifications, kafka.TopicSpecification{
			Topic:             param.Topic,
			NumPartitions:     NumPartitions,
			ReplicationFactor: ReplicationFactor,
		})
	}

	// Cria os tópicos (somente os que não existem)
	ctx, cancel := context.WithTimeout(context.Background(), adminTimeout)
	defer cancel()

	results, err := adminClient.CreateTopics(ctx, topicSpecifications)
	if err != nil {
		log.Printf("Erro ao criar tópicos: %v", err)
		return
	}

	// Log o status da criação
	for _, result := range results {
		if result.Error.Code() == kafka.ErrTopicAlreadyExists {
			continue
		}
		if result.Error.Code() != kafka.ErrNoError {
			log.Printf("Erro ao criar tópico %s: %v\n", result.Topic, result.Error)
		} else {
			log.Printf("Tópico criado com sucesso: %s\n", result.Topic)
		}
	}
}

// readTopics consome os tópicos até o context ser cancelado.
func readTopics(ctx context.Context, topicParams []KafkaReadTopicsParams) {
	if len(topicParams) == 0 {
		log.Println("Nenhum tópico para consumir")
		return
	}

	topics := make([]string, len(topicParams))
	for i, topicParam := range topicParams {
		topics[i] = topicParam.Topic
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        KafkaBootstrapServers,
		"group.id":                 KafkaGroupID,
		"client.id":                KafkaClientID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"session.timeout.ms":       6000,
		"reconnect.backoff.ms":     50,
		"reconnect.backoff.max.ms": 1000,
	})
	if err != nil {
		// Sem o return, o defer abaixo rodaria sobre um ponteiro nil.
		log.Printf("Erro ao criar o consumidor: %v", err)
		return
	}
	defer func() {
		if closeErr := consumer.Close(); closeErr != nil {
			log.Printf("Erro ao fechar o consumidor: %v", closeErr)
		}
	}()

	if err := consumer.SubscribeTopics(topics, nil); err != nil {
		log.Printf("Erro ao subscrever-se aos tópicos: %v", err)
		return
	}

	log.Println("Consumidor iniciado. Aguardando mensagens...")

	// Mapeia os handlers para os tópicos
	handlerMap := make(map[string]func(*kafka.Message) error, len(topicParams))
	for _, param := range topicParams {
		handlerMap[param.Topic] = param.Handler
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumidor encerrado")
			return
		default:
		}

		msg, err := consumer.ReadMessage(readPollTimeout)
		if err != nil {
			// Timeout é o caminho normal: só devolve o controle para o select.
			var kafkaErr kafka.Error
			if errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrTimedOut {
				continue
			}

			log.Printf("Erro ao ler mensagem: %v\n", err)
			continue
		}

		// Recuperar o handler associado ao tópico
		handler, exists := handlerMap[*msg.TopicPartition.Topic]
		if !exists {
			log.Printf("Nenhum handler encontrado para o tópico: %s\n", *msg.TopicPartition.Topic)
			continue
		}

		// Processar a mensagem usando o handler
		if err := handler(msg); err != nil {
			log.Printf("Erro ao processar mensagem do tópico %s: %v\n", *msg.TopicPartition.Topic, err)
			continue
		}

		// Commit manual após sucesso
		if _, commitErr := consumer.CommitMessage(msg); commitErr != nil {
			log.Printf("Erro ao fazer commit do offset: %v", commitErr)
		} else {
			log.Printf("Leitura confirmada com sucesso para o tópico %s", *msg.TopicPartition.Topic)
		}
	}
}

func PublishMessage(topic string, message string) error {
	if producer == nil {
		return errors.New("kafka producer não inicializado")
	}

	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, deliveryChan)
	if err != nil {
		return err
	}

	e := <-deliveryChan
	m, ok := e.(*kafka.Message)
	if !ok {
		return errors.New("evento de entrega inesperado")
	}

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}

	log.Printf("Delivered message to %v", m.TopicPartition)

	return nil
}
