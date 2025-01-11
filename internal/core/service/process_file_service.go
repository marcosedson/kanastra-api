package service

type DebtRepository interface {
	Save(hashDebtID string) error
	IsLineProcessed(hashDebtID string) bool
}
