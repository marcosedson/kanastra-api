package usecase

import (
	"bytes"
	"errors"
	"kanastra-api/internal/core/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	MockDebtRepository struct {
		mock.Mock
	}

	MockEmailPublisher struct {
		mock.Mock
	}

	MockInvoiceGenerator struct {
		mock.Mock
	}

	MockKafkaProducer struct {
		mock.Mock
	}
)

func (m *MockDebtRepository) Save(debtID string) error {
	args := m.Called(debtID)

	return args.Error(0)
}

func (m *MockDebtRepository) IsLineProcessed(debtID string) bool {
	args := m.Called(debtID)

	return args.Bool(0)
}

func (m *MockEmailPublisher) Publish(email string, debt domain.Debt) {
	m.Called(email, debt)
}

func (m *MockInvoiceGenerator) Generate(debt domain.Debt) {
	m.Called(debt)
}

func (m *MockKafkaProducer) Produce(key string, value []byte) error {
	args := m.Called(key, value)

	return args.Error(0)
}

func TestProcessFileAsync_Success(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)

	fileContent := `Name,GovernmentID,Email,DebtAmount,DebtDueDate,DebtID
John Doe,1234,john.doe@example.com,100.00,2025-01-01,1a2b3c4d
Jane Doe,5678,jane.doe@example.com,200.50,2025-02-02,2a2b3c4d`

	repo.On("IsLineProcessed", "1a2b3c4d").Return(false)
	repo.On("IsLineProcessed", "2a2b3c4d").Return(false)
	repo.On("Save", "1a2b3c4d").Return(nil)
	repo.On("Save", "2a2b3c4d").Return(nil)

	producer.On("Produce", mock.Anything, mock.Anything).Return(nil)

	totalLines := useCase.ProcessFileAsync(bytes.NewReader([]byte(fileContent)), "test.csv")

	assert.Equal(t, 2, totalLines)
	repo.AssertNumberOfCalls(t, "Save", 2)
	producer.AssertNumberOfCalls(t, "Produce", 2)
}

func TestProcessFileAsync_EmptyFile(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)

	fileContent := ``
	totalLines := useCase.ProcessFileAsync(bytes.NewReader([]byte(fileContent)), "test.csv")

	assert.Equal(t, 0, totalLines)
	repo.AssertNotCalled(t, "Save", mock.Anything)
	producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything)
}

func TestProcessFileAsync_HeaderError(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)

	fileContent := `Name,GovernmentID,Email,DebtAmount,DebtDueDate`

	totalLines := useCase.ProcessFileAsync(bytes.NewReader([]byte(fileContent)), "test.csv")

	assert.Equal(t, 0, totalLines)
	repo.AssertNotCalled(t, "Save", mock.Anything)
	producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything)
}

func TestProcessFileAsync_LastBatchError(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)

	fileContent := `Name,GovernmentID,Email,DebtAmount,DebtDueDate,DebtID
John Doe,1234,john.doe@example.com,100.00,2025-01-01,1a2b3c4d`

	producer.On("Produce", mock.Anything, mock.Anything).Return(errors.New("erro ao produzir mensagem"))
	repo.On("IsLineProcessed", "1a2b3c4d").Return(false)

	totalLines := useCase.ProcessFileAsync(bytes.NewReader([]byte(fileContent)), "test.csv")

	assert.Equal(t, 1, totalLines)
	producer.AssertCalled(t, "Produce", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything)
}

func TestSendBatch_EmptyBatch(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)
	err := useCase.sendBatch("test.csv", [][]string{})
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "Save", mock.Anything)
	producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything)
}

func TestSendBatch_DuplicateLine(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)

	record := []string{"John Doe", "1234", "john.doe@example.com", "100.00", "2025-01-01", "1a2b3c4d"}
	batch := [][]string{record}

	repo.On("IsLineProcessed", "1a2b3c4d").Return(true)

	err := useCase.sendBatch("test.csv", batch)
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "Save", "1a2b3c4d")
	producer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything)
}

func TestSendBatch_SaveError(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)

	record := []string{"John Doe", "1234", "john.doe@example.com", "100.00", "2025-01-01", "1a2b3c4d"}
	batch := [][]string{record}

	repo.On("IsLineProcessed", "1a2b3c4d").Return(false)
	repo.On("Save", "1a2b3c4d").Return(errors.New("erro ao salvar no repositório"))

	producer.On("Produce", mock.Anything, mock.Anything).Return(nil)

	err := useCase.sendBatch("test.csv", batch)
	assert.Error(t, err)
	repo.AssertCalled(t, "Save", "1a2b3c4d")
	producer.AssertCalled(t, "Produce", "test.csv", mock.Anything)
}

func TestSendBatch_ProduceError(t *testing.T) {
	repo := new(MockDebtRepository)
	email := new(MockEmailPublisher)
	invoice := new(MockInvoiceGenerator)
	producer := new(MockKafkaProducer)

	useCase := NewProcessFileUseCase(repo, email, invoice, producer)

	record := []string{"John Doe", "1234", "john.doe@example.com", "100.00", "2025-01-01", "1a2b3c4d"}
	batch := [][]string{record}

	repo.On("IsLineProcessed", "1a2b3c4d").Return(false)
	producer.On("Produce", "test.csv", mock.Anything).Return(errors.New("erro ao enviar mensagem"))

	err := useCase.sendBatch("test.csv", batch)
	assert.Error(t, err)
	producer.AssertCalled(t, "Produce", "test.csv", mock.Anything)
	repo.AssertNotCalled(t, "Save", "1a2b3c4d")
}

func TestValidators(t *testing.T) {
	assert.True(t, IsValidGovernmentID("12345"))
	assert.False(t, IsValidGovernmentID("abc"))

	assert.True(t, IsValidEmail("test@example.com"))
	assert.False(t, IsValidEmail("invalid-email"))

	assert.True(t, IsValidDebtAmount("100.50"))
	assert.False(t, IsValidDebtAmount("abc"))

	assert.True(t, IsValidDebtDueDate("2025-12-31"))
	assert.False(t, IsValidDebtDueDate("31/12/2025"))
}
