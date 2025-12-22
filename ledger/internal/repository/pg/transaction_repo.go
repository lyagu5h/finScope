package pg

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func (r TransactionRepository) Add(ctx context.Context, tx *domain.Transaction) error {
	const q = `INSERT INTO expenses (amount, category, description, date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err := r.db.QueryRowContext(
		ctx,
		q,
		tx.Amount,
		tx.Category,
		tx.Description,
		tx.Date,
	).Scan(&tx.ID)
	
	return err
}

func (r TransactionRepository) List(ctx context.Context) ([]domain.Transaction, error) {
	const q = `
		SELECT id, amount, category, description, date
		FROM expenses
		ORDER BY date DESC, id DESC
	`
	
	rows, err := r.db.QueryContext(ctx, q)
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

func (r TransactionRepository) SumByCategory(ctx context.Context, category string) (float64, error) {
	var sum sql.NullFloat64
	const q = `
		SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1
	`
	if err := r.db.QueryRowContext(
		ctx,
		q,
		category,
		).Scan(&sum); err != nil {
			log.Println("DB ERROR:", err)
			return 0, err
		}

		if !sum.Valid {
			return 0, nil
		}

	return sum.Float64, nil
}

func (r TransactionRepository) ListCategories(ctx context.Context) ([]string, error) {
	const q = `
		SELECT DISTINCT category FROM expenses
		ORDER BY category ASC
	`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r TransactionRepository) SumByCategoryAndPeriod(
	ctx context.Context,
	category string,
	from, to time.Time,
) (float64, error) {
	var sum sql.NullFloat64

	const q = `
		SELECT COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE category = $1
		  AND date >= $2
		  AND date <= $3
	`

	if err := r.db.QueryRowContext(
		ctx,
		q,
		category,
		from,
		to,
	).Scan(&sum); err != nil {
		return 0, err
	}

	if !sum.Valid {
		return 0, nil
	}

	return sum.Float64, nil
}