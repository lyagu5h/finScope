package pg

import (
	"database/sql"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func (r TransactionRepository) Add(tx *domain.Transaction) error {
	return r.db.QueryRow(`
		INSERT INTO expenses (amount, category, description, date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		tx.Amount,
		tx.Category,
		tx.Description,
		tx.Date,
	).Scan(&tx.ID)
}

func (r TransactionRepository) List() ([]domain.Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, amount, category, description, date
		FROM expenses
		ORDER BY date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.Amount,
			&tx.Category,
			&tx.Description,
			&tx.Date,
		); err != nil {
			return nil, err
		}
		res = append(res, tx)
	}

	return res, rows.Err()
}

func (r TransactionRepository) SumByCategory(category string) (float64, error) {
	var sum sql.NullFloat64
	if err := r.db.QueryRow(
			`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1`,
			category,
		).Scan(&sum); err != nil {
			return 0, err
		}

		if !sum.Valid {
			return 0, nil
		}
	return sum.Float64, nil
}