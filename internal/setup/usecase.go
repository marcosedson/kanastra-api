package setup

import (
	"kanastra-api/internal/core/usecase"
	"kanastra-api/internal/infra/adapter/external"
	"kanastra-api/internal/infra/adapter/kafka"
	"kanastra-api/internal/infra/adapter/persistence"
)

func UseCase(
	repo *persistence.DebtRepository,
	email *external.EmailPublisher,
	invoice *external.InvoiceGenerator,
	producer *kafka.DynamicProducer,
) *usecase.ProcessFileUseCase {
	return usecase.NewProcessFileUseCase(repo, email, invoice, producer)
}
