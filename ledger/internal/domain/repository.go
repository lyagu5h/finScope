package domain

type BudgetRepository interface {
	Upsert(b Budget) error
	GetByCategory(category string) (Budget, error)
	List() ([]Budget, error)
}

type TransactionRepository interface {
	Add(tx *Transaction) error
	List() ([]Transaction, error)
	SumByCategory(category string) (float64, error)
}