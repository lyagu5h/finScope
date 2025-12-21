package service

import (
	"errors"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
)

var ErrBudgetExceeded = errors.New("budget exceeded")


type LedgerService interface {
	SetBudget( b domain.Budget) error
	ListBudgets() ([]domain.Budget, error)

	AddTransaction(t domain.Transaction) (domain.Transaction, error) 
	ListTransactions() ([]domain.Transaction, error)
}

type ledger struct {
	budgets domain.BudgetRepository
	transactions domain.TransactionRepository
}

func New(
	budgetsRepo domain.BudgetRepository,
	transactionsRepo domain.TransactionRepository,
) LedgerService {
	return &ledger{
		budgets:      budgetsRepo,
		transactions: transactionsRepo,
	}
}


//NEED TO FIX
func (svc *ledger) AddTransaction(t domain.Transaction) (domain.Transaction, error) {
	if err := t.Validate(); err != nil {
		return t, err
	}

	budget, err := svc.budgets.GetByCategory(t.Category)
	if err != nil {
		return t, err
	}

	current, err := svc.transactions.SumByCategory(t.Category)
	if err != nil {
		return t, err
	}

	if current+t.Amount > budget.Limit {
		return t, ErrBudgetExceeded
	}

	if err := svc.transactions.Add(&t); err != nil {
		return t, err
	}

	return t, nil
}

func (svc *ledger) ListTransactions() ([]domain.Transaction, error) {
	return svc.transactions.List()
}

func (svc *ledger) SetBudget(b domain.Budget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	return svc.budgets.Upsert(b)
}

func (svc *ledger) ListBudgets() ([]domain.Budget, error) {
	return svc.budgets.List()
}