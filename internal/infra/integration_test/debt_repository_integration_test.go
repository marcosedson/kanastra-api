package integration_test

import (
	"testing"

	"kanastra-api/internal/infra/adapter/persistence"

	"github.com/stretchr/testify/assert"
)

func TestDebtRepositoryIntegration(t *testing.T) {
	repo := persistence.NewDebtRepository()

	err := repo.Save("12345")
	assert.NoError(t, err, "Erro ao salvar o ID pela primeira vez")

	isProcessed := repo.IsLineProcessed("12345")
	assert.True(t, isProcessed, "O ID salvo deveria ser processado")

	err = repo.Save("12345")
	assert.NoError(t, err, "O repositório deve permitir salvar IDs repetidos sem falhas")
}
