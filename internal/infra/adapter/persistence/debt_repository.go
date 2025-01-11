package persistence

import (
	"sync"
)

type DebtRepository struct {
	store map[string]struct{}
	mu    sync.Mutex
}

func NewDebtRepository() *DebtRepository {
	return &DebtRepository{
		store: make(map[string]struct{}),
	}
}

func (r *DebtRepository) Save(debtID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[debtID] = struct{}{}
	return nil
}

func (r *DebtRepository) IsLineProcessed(debtID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.store[debtID]
	return exists
}
