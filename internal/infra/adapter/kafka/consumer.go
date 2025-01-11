package kafka

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"kanastra-api/internal/core/domain"
)

type Reader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type DebtRepositoryInterface interface {
	Save(debtID string) error
}

type Consumer struct {
	reader         Reader
	DebtRepository DebtRepositoryInterface
}

func NewKafkaConsumer(brokerAddress, topic, groupID string, repo DebtRepositoryInterface) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{brokerAddress},
		Topic:          topic,
		GroupID:        groupID,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
	})
	return &Consumer{
		reader:         reader,
		DebtRepository: repo,
	}
}

func (c *Consumer) Consume(processMessage func(debt domain.Debt, fileName string)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageChan := make(chan kafka.Message, 1000)
	defer close(messageChan)

	numWorkers := 10

	for i := 0; i < numWorkers; i++ {
		go func() {
			for message := range messageChan {
				fileName := string(message.Key)
				reader := csv.NewReader(strings.NewReader(string(message.Value)))
				record, err := reader.Read()
				if err != nil {
					log.Printf("Erro ao processar CSV: %v, Mensagem: %s", err, string(message.Value))
					continue
				}

				debt := domain.Debt{
					Name:         record[0],
					GovernmentID: record[1],
					Email:        record[2],
					DebtAmount:   parseFloat(record[3]),
					DebtDueDate:  record[4],
					DebtID:       record[5],
				}

				processMessage(debt, fileName)

				if err := c.DebtRepository.Save(debt.DebtID); err != nil {
					log.Printf("Erro ao salvar no repositório: %v", err)
					continue
				}

				if err := c.reader.CommitMessages(ctx, message); err != nil {
					log.Printf("Erro ao confirmar mensagem: %v", err)
				}
			}
		}()
	}

	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				log.Println("Contexto encerrado ou timeout atingido, encerrando loop.")

				break
			}

			if err == io.EOF {
				break
			}

			log.Printf("Erro ao consumir mensagem: %v", err)

			continue
		}

		messageChan <- message
	}

	return nil
}

func parseFloat(value string) float64 {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Printf("Erro ao converter string para float: %v", err)

		return 0
	}

	return parsedValue
}

func (c *Consumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Printf("Erro ao fechar o consumer Kafka: %v", err)
	}
}
