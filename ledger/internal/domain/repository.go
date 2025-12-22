package domain

import (
	"context"
	"time"
)

type BudgetRepository interface {
	Upsert(ctx context.Context, b Budget) error
	GetByCategory(ctx context.Context, category string) (Budget, bool, error)
	List(ctx context.Context) ([]Budget, error)
}

type TransactionRepository interface {
	Add(ctx context.Context, tx *Transaction) error
	List(ctx context.Context) ([]Transaction, error)
	SumByCategory(ctx context.Context, category string) (float64, error)
	ListCategories(ctx context.Context) ([]string, error)

	SumByCategoryAndPeriod(
		ctx context.Context,
		category string,
		from, to time.Time,
	) (float64, error)
}