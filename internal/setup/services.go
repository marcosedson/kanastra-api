package setup

import (
	"kanastra-api/internal/infra/adapter/external"
)

func Services() (*external.EmailPublisher, *external.InvoiceGenerator) {
	email := external.NewEmailPublisher()
	invoice := external.NewInvoiceGenerator()

	return email, invoice
}
