package external

import (
	"log"

	"kanastra-api/internal/core/domain"
)

type EmailPublisher struct{}

func NewEmailPublisher() *EmailPublisher {
	return &EmailPublisher{}
}

func (e *EmailPublisher) Publish(email string, debt domain.Debt) {
	log.Printf("E-mail enviado com sucesso para %s sobre débito: %+v", email, debt)
}
