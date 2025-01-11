package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebtRepository_Save(t *testing.T) {
	repo := NewDebtRepository()

	t.Run("Save new debtID", func(t *testing.T) {
		debtID := "12345"
		err := repo.Save(debtID)
		assert.NoError(t, err)

		assert.True(t, repo.IsLineProcessed(debtID))
	})

	t.Run("Save duplicate debtID", func(t *testing.T) {
		debtID := "12345"
		err := repo.Save(debtID)
		assert.NoError(t, err)

		assert.True(t, repo.IsLineProcessed(debtID))
	})
}

func TestDebtRepository_IsLineProcessed(t *testing.T) {
	repo := NewDebtRepository()

	t.Run("Check non-existent debtID", func(t *testing.T) {
		debtID := "99999"
		assert.False(t, repo.IsLineProcessed(debtID))
	})

	t.Run("Check existent debtID", func(t *testing.T) {
		debtID := "67890"
		err := repo.Save(debtID)
		assert.NoError(t, err)

		assert.True(t, repo.IsLineProcessed(debtID))
	})
}
