package setup

import (
	"kanastra-api/internal/infra/adapter/persistence"
)

func Repository() *persistence.DebtRepository {
	return persistence.NewDebtRepository()
}
