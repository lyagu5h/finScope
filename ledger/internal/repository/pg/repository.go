package pg

import (
	"database/sql"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
)

type Repositories struct {
	BudgetRepository      domain.BudgetRepository
	TransactionRepository domain.TransactionRepository
}

func New(db *sql.DB) *Repositories {
	return &Repositories{
		BudgetRepository:      BudgetRepository{db: db},
		TransactionRepository: TransactionRepository{db: db},
	}
}