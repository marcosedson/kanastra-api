package external

import (
	"bytes"
	"kanastra-api/internal/core/domain"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailPublisher_Publish(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	emailPublisher := NewEmailPublisher()

	tests := []struct {
		name  string
		email string
		debt  domain.Debt
	}{
		{
			name:  "Caso com e-mail e débito válidos",
			email: "test@example.com",
			debt: domain.Debt{
				Name:         "John Doe",
				GovernmentID: "1234567890",
				Email:        "test@example.com",
				DebtAmount:   1200.50,
				DebtDueDate:  "2024-12-31",
				DebtID:       "abc123",
			},
		},
		{
			name:  "Caso com valores vazios",
			email: "",
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
			name:  "Caso com dados inconsistentes",
			email: "invalidemail@",
			debt: domain.Debt{
				Name:         "Nome Inválido",
				GovernmentID: "12",
				Email:        "invalidemail@",
				DebtAmount:   -500.00,
				DebtDueDate:  "31/02/2025",
				DebtID:       "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()

			emailPublisher.Publish(tt.email, tt.debt)

			assert.Contains(t, logBuffer.String(), "E-mail enviado com sucesso para", "O log deve conter a mensagem de envio.")
			assert.Contains(t, logBuffer.String(), tt.email, "O log deve conter o e-mail fornecido.")
		})
	}
}

func TestNewEmailPublisher(t *testing.T) {
	emailPublisher := NewEmailPublisher()

	assert.NotNil(t, emailPublisher, "NewEmailPublisher() deve retornar uma instância não nula")
	assert.IsType(t, &EmailPublisher{}, emailPublisher, "NewEmailPublisher() deve retornar uma instância do tipo EmailPublisher")
}
