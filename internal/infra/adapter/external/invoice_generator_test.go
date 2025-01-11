package external

import (
	"bytes"
	"kanastra-api/internal/core/domain"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvoiceGenerator_Generate(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	invoiceGenerator := NewInvoiceGenerator()

	tests := []struct {
		name string
		debt domain.Debt
	}{
		{
			name: "Caso com Debt completo e válido",
			debt: domain.Debt{
				Name:         "João Silva",
				GovernmentID: "987654321",
				Email:        "joao.silva@example.com",
				DebtAmount:   500.75,
				DebtDueDate:  "2023-12-31",
				DebtID:       "001",
			},
		},
		{
			name: "Caso com campos vazios no Debt",
			debt: domain.Debt{
				Name:         "",
				GovernmentID: "",
				Email:        "",
				DebtAmount:   0.0,
				DebtDueDate:  "",
				DebtID:       "",
			},
		},
		{
			name: "Caso com dados incorretos/inconsistentes no Debt",
			debt: domain.Debt{
				Name:         "Fulano de Tal",
				GovernmentID: "12",
				Email:        "emailinvalido@",
				DebtAmount:   -100.0,
				DebtDueDate:  "2023/12/31",
				DebtID:       "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()

			invoiceGenerator.Generate(tt.debt)

			assert.Contains(t, logBuffer.String(), "Boleto gerado com sucesso para o débito:", "Log deve indicar que o boleto foi gerado com sucesso.")
			assert.Contains(t, logBuffer.String(), tt.debt.Name, "Log deve conter o nome do cliente.")
			assert.Contains(t, logBuffer.String(), tt.debt.DebtID, "Log deve conter o ID do débito.")
		})
	}
}

func TestNewInvoiceGenerator(t *testing.T) {
	invoiceGenerator := NewInvoiceGenerator()

	assert.NotNil(t, invoiceGenerator, "NewInvoiceGenerator() deve retornar uma instância não nula")
	assert.IsType(t, &InvoiceGenerator{}, invoiceGenerator, "NewInvoiceGenerator() deve retornar uma instância do tipo InvoiceGenerator")
}
