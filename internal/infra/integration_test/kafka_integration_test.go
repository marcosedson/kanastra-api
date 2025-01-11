package integration_test

import (
	"kanastra-api/internal/setup"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"kanastra-api/internal/core/domain"
	"kanastra-api/internal/infra/adapter/kafka"
)

type mockDebtRepository struct{}

func (m *mockDebtRepository) Save(debtID string) error {
	log.Printf("Mock repository: dívida salva com ID %s", debtID)

	return nil
}

type mockEmailPublisher struct{}

func (m *mockEmailPublisher) Publish(email string, _ domain.Debt) error {
	log.Printf("Mock email: enviado para %s com sucesso.", email)

	return nil
}

type mockInvoiceGenerator struct{}

func (m *mockInvoiceGenerator) Generate(debt domain.Debt) error {
	log.Printf("Mock fatura: gerada para a dívida %+v", debt)
	return nil
}

func TestKafkaSetupIntegration(t *testing.T) {
	broker := os.Getenv("BROKER_ADDRESS")
	if broker == "" {
		broker = "localhost:9092"
	}

	kafka.WaitForKafka(broker, 30*time.Second)

	mockRepo := &mockDebtRepository{}

	producer, consumer := setup.Kafka(mockRepo)
	defer setup.CloseKafka(producer, consumer)

	externalEmail := &mockEmailPublisher{}
	externalInvoice := &mockInvoiceGenerator{}

	go func() {
		consumerErr := consumer.Consume(func(debt domain.Debt, fileName string) {
			log.Printf("Mensagem recebida: %+v", debt)

			err := externalInvoice.Generate(debt)
			assert.NoError(t, err, "Erro ao gerar fatura no mock")

			err = externalEmail.Publish(debt.Email, debt)
			assert.NoError(t, err, "Erro ao enviar email no mock")

			log.Printf("Mensagem processada com sucesso no teste: %+v", debt)
		})
		assert.NoError(t, consumerErr, "Erro no consumidor Kafka durante o consumo")
	}()

	testMessage := []byte("John Doe,123456789,johndoe@example.com,200.5,2023-12-31,debtID-123")
	err := producer.Produce("debt-123", testMessage)
	assert.NoError(t, err, "Erro ao produzir mensagem para o Kafka")

	producer.ProcessQueue()

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: Mensagem não processada no teste de integração Kafka")
	default:
		log.Println("Teste de integração Kafka finalizado com sucesso")
	}
}
