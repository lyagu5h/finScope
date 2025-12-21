package pg

import (
	"database/sql"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
)


type BudgetRepository struct {
	db *sql.DB
}

func (r BudgetRepository) Upsert(b domain.Budget) error {
	_, err := r.db.Exec(
		`INSERT INTO budgets (category, limit_amount)
		 VALUES ($1, $2)
		 ON CONFLICT (category)
		 DO UPDATE SET limit_amount = EXCLUDED.limit_amount`,
		b.Category, b.Limit,
	)
	return err
}

func (r BudgetRepository) GetByCategory(category string) (domain.Budget, error) {
	var b domain.Budget
	err := r.db.QueryRow(
		`SELECT category, limit FROM budgets WHERE category = $1`,
		category,
	).Scan(&b.Category, &b.Limit)

	if err == sql.ErrNoRows {
		return domain.Budget{}, nil
	}

	if err != nil {
		return domain.Budget{}, err
	}

	return b, nil
}

func (r BudgetRepository) List() ([]domain.Budget, error) {
	const q = `
		SELECT category, limit_amount
		FROM budgets
		ORDER BY category
	`

	rows, err := r.db.Query(q)

	if err != nil {
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