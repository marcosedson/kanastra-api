package domain

type Debt struct {
	Name         string  `json:"Name"`
	GovernmentID string  `json:"GovernmentID"`
	Email        string  `json:"Email"`
	DebtAmount   float64 `json:"DebtAmount"`
	DebtDueDate  string  `json:"DebtDueDate"`
	DebtID       string  `json:"DebtID"`
}
