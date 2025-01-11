package setup

import (
	"log"
	"time"

	"kanastra-api/internal/core/domain"
	"kanastra-api/internal/infra/adapter/external"
	"kanastra-api/internal/infra/adapter/kafka"
	"kanastra-api/internal/infra/config"
)

func Kafka(repo kafka.DebtRepositoryInterface) (*kafka.DynamicProducer, *kafka.Consumer) {
	broker := config.GetEnv("BROKER_ADDRESS", "localhost:9092")
	topic := config.GetEnv("TOPIC", "default_topic")
	groupID := config.GetEnv("GROUP_ID", "default_group")

	kafka.WaitForKafka(broker, 60*time.Second)

	producer := kafka.NewDynamicKafkaProducer(broker, topic)
	consumer := kafka.NewKafkaConsumer(broker, topic, groupID, repo)

	go startKafkaConsumer(consumer, external.NewEmailPublisher(), external.NewInvoiceGenerator())

	return producer, consumer
}

func CloseKafka(producer *kafka.DynamicProducer, consumer *kafka.Consumer) {
	producer.Close()
	consumer.Close()
}

func startKafkaConsumer(consumer *kafka.Consumer, email *external.EmailPublisher, invoice *external.InvoiceGenerator) {
	err := consumer.Consume(func(debt domain.Debt, filename string) {
		log.Printf("Mensagem recebida: %+v", debt)

		invoice.Generate(debt)
		email.Publish(debt.Email, debt)

		log.Printf("Mensagem processada com sucesso: %+v", debt)
	})

	if err != nil {
		log.Printf("Erro no consumidor Kafka: %v", err)
	}
}
