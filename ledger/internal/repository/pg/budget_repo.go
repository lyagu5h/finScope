package pg

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
)


type BudgetRepository struct {
	db *sql.DB
	log *slog.Logger
}

func (r BudgetRepository) Upsert(ctx context.Context,b domain.Budget) error {
	const q = `INSERT INTO budgets (category, limit_amount)
		 VALUES ($1, $2)
		 ON CONFLICT (category)
		 DO UPDATE SET limit_amount = EXCLUDED.limit_amount`
	_, err := r.db.ExecContext(
		ctx,
		q,
		b.Category, b.Limit,
	)
	return err
}

func (r BudgetRepository) GetByCategory(ctx context.Context, category string) (domain.Budget, bool, error) {
	var b domain.Budget
	const q = `SELECT category, limit_amount FROM budgets WHERE category = $1`
	err := r.db.QueryRowContext(
		ctx,
		q,
		category,
	).Scan(&b.Category, &b.Limit)

	if err == sql.ErrNoRows {
		return domain.Budget{}, false, nil
	}

	if err != nil {
		if errors.Is(err, context.Canceled) {
			r.log.Warn("db query cancelled")
		}

		log.Println("errror:", err)
		return domain.Budget{}, false, err
	}

	return b, true, nil
}

func (r BudgetRepository) List(ctx context.Context) ([]domain.Budget, error) {
	const q = `
		SELECT category, limit_amount
		FROM budgets
		ORDER BY category
	`

	rows, err := r.db.QueryContext(ctx, q)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			r.log.Warn("db query cancelled")
		} else {
			r.log.Error("db query failed", slog.String("error", err.Error()))
		}
		return nil, err
	}

	defer rows.Close()

	var res []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return nil, err
		}
		res = append(res, b)
	}

	return res, rows.Err()
}