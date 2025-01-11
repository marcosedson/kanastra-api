package usecase

import (
	"encoding/csv"
	"fmt"
	"io"
	"kanastra-api/internal/core/domain"
	"kanastra-api/internal/core/service"
	"log"
	"regexp"
	"strings"
)

type EmailPublisher interface {
	Publish(email string, debt domain.Debt)
}

type InvoiceGenerator interface {
	Generate(debt domain.Debt)
}

type KafkaProducer interface {
	Produce(key string, value []byte) error
}

type ProcessFileUseCase struct {
	repo     service.DebtRepository
	email    EmailPublisher
	invoice  InvoiceGenerator
	producer KafkaProducer
}

func NewProcessFileUseCase(repo service.DebtRepository, email EmailPublisher, invoice InvoiceGenerator, producer KafkaProducer) *ProcessFileUseCase {
	return &ProcessFileUseCase{repo: repo, email: email, invoice: invoice, producer: producer}
}

func (u *ProcessFileUseCase) ProcessFileAsync(file io.Reader, fileName string) (totalLines int) {
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = -1
	batchSize := 1000
	var batch [][]string

	_, err := reader.Read()
	if err != nil {
		log.Printf("Erro ao ignorar o cabeçalho do arquivo: %v", err)

		return
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Printf("Erro ao ler linha do arquivo: %v", err)
			continue
		}

		totalLines++

		batch = append(batch, record)
		if len(batch) == batchSize {
			if err := u.sendBatch(fileName, batch); err != nil {
				log.Printf("Erro ao enviar lote para o Kafka (arquivo: %s): %v", fileName, err)

				continue
			}
			batch = nil
		}
	}

	if len(batch) > 0 {
		if err := u.sendBatch(fileName, batch); err != nil {
			log.Printf("Erro ao enviar último lote para o Kafka (arquivo: %s): %v", fileName, err)
		}
	}

	return totalLines
}

func (u *ProcessFileUseCase) sendBatch(fileName string, batch [][]string) error {
	for _, record := range batch {
		err := validateRecord(record)
		if err != nil {
			log.Printf("Linha inválida: %v, Erro: %v", record, err)

			continue
		}

		if u.repo.IsLineProcessed(record[5]) {
			log.Printf("Linha já foi processada: %v", record)

			continue
		}

		message := strings.Join(record, ",")
		if err := u.producer.Produce(fileName, []byte(message)); err != nil {
			log.Printf("Erro ao enviar mensagem ao Kafka: %v", err)

			return err
		}

		err = u.repo.Save(record[5])
		if err != nil {
			return err
		}

		log.Printf("Mensagem enviada ao Kafka com sucesso: %s", message)
	}

	return nil
}

func validateRecord(record []string) error {
	if len(record) != 6 {
		return fmt.Errorf("registro inválido: número incorreto de campos")
	}

	if !IsValidGovernmentID(record[1]) {
		return fmt.Errorf("governmentID inválido: %s", record[1])
	}

	if !IsValidEmail(record[2]) {
		return fmt.Errorf("email inválido: %s", record[2])
	}

	if !IsValidDebtAmount(record[3]) {
		return fmt.Errorf("debtAmount inválido: %s", record[3])
	}

	if !IsValidDebtDueDate(record[4]) {
		return fmt.Errorf("debtDueDate inválida: %s", record[4])
	}

	return nil
}

func IsValidGovernmentID(governmentID string) bool {
	regexPattern := `^\d{4,11}$`
	matched, _ := regexp.MatchString(regexPattern, governmentID)

	return matched
}

func IsValidDebtDueDate(date string) bool {
	regexPattern := `^\d{4}-\d{2}-\d{2}$`
	matched, _ := regexp.MatchString(regexPattern, date)

	return matched
}

func IsValidDebtAmount(amount string) bool {
	regexPattern := `^\d+(\.\d{1,2})?$`
	matched, _ := regexp.MatchString(regexPattern, amount)

	return matched
}

func IsValidEmail(email string) bool {
	regexPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(regexPattern, email)

	return matched
}
