package external

import (
	"log"

	"kanastra-api/internal/core/domain"
)

type InvoiceGenerator struct{}

func NewInvoiceGenerator() *InvoiceGenerator {
	return &InvoiceGenerator{}
}

func (b InvoiceGenerator) Generate(debt domain.Debt) {
	log.Printf("Boleto gerado com sucesso para o débito: %+v", debt)
}
