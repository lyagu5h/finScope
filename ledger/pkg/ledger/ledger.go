package ledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/lyagu5h/finScope/ledger/pkg/db"
)

//errors
var ErrBudgetExceeded = errors.New("budget exceeded")
var ErrBudgetCategoryEmpty = errors.New("budget category cannot be empty")
var ErrBudgetLimitZero = errors.New("budget limit should be > 0")
var ErrTransactionAmountZero = errors.New("transaction amount should be > 0")
var ErrTransactionCategoryEmpty = errors.New("transaction category cannot be empty")
var ErrTransactionDateEmpty = errors.New("transaction date must be set")

type Validatable interface {
	Validate() error
}

type Budget struct {
	Category string `json:"category"`
	Limit float64 `json:"limit"`
	Period string `json:"period,omitempty"`
}

func (b Budget) Validate() error {
	if b.Category == "" {
		return ErrBudgetCategoryEmpty
	}
	if b.Limit <= 0 {
		return ErrBudgetLimitZero
	}

	return nil
}

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

func (tx Transaction) Validate() error {
	
	if tx.Amount == 0 || tx.Amount < 0 {
		return ErrTransactionAmountZero
	}

	if tx.Category == "" {
		return ErrTransactionCategoryEmpty
	}

	if tx.Date.IsZero() {
		return ErrTransactionDateEmpty
	}

	return nil
}


type Store struct {
	// db  *sql.DB	
	logger *slog.Logger
}

func NewStore(log *slog.Logger) *Store {
	return &Store{
		logger: log,
	}
}

func (s *Store) SetBudget(b Budget) error {

	err := b.Validate()
	if err != nil {
		return fmt.Errorf("SetBudget: %s", err)
	}

	const q = `
		INSERT INTO budgets (category, limit_amount)
		VALUES ($1, $2)
		ON CONFLICT (category)
		DO UPDATE SET limit_amount = EXCLUDED.limit_amount
	`

	_, err = db.DB.Exec(q, b.Category, b.Limit)
	
	return err
}

func (s *Store) ListBudgets() ([]Budget, error) {
	const q = `
		SELECT caZtegory, limit_amount
		FROM budgets
		ORDER BY category
	`

	rows, err := db.DB.Query(q)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return nil, err
		}
		res = append(res, b)
	}

	return res, rows.Err()
}

func (s *Store) LoadBudgets(r io.Reader) error {
	var budgets []Budget
	dec := json.NewDecoder(r)

	if err := dec.Decode(&budgets); err != nil {
		return err
	}

	for i, b := range budgets {
		if err := s.SetBudget(b); err != nil {
			return fmt.Errorf("LoadBudget: %s at %d, %v ", err, i, b)
		}
	}

	return nil

}

func (s *Store) AddTransaction(tx Transaction) error {

	if err := tx.Validate(); err != nil {
		return err
	}
	var limit float64
	err := db.DB.QueryRow(
		`SELECT limit_amount FROM budgets WHERE category = $1`,
		tx.Category,
	).Scan(&limit)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		var spent float64
		if err := db.DB.QueryRow(
			`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1`,
			tx.Category,
		).Scan(&spent); err != nil {
			return err
		}

		if spent+tx.Amount > limit {
			return ErrBudgetExceeded
		}
	}
	const q = `
		INSERT INTO expenses (amount, category, description, date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	return db.DB.QueryRow(
		q,
		tx.Amount,
		tx.Category,
		tx.Description,
		tx.Date,
	).Scan(&tx.ID)
}

func (s *Store) ListTransactions() ([]Transaction, error) {
const q = `
		SELECT id, amount, category, description, date
		FROM expenses
		ORDER BY date DESC, id DESC
	`

	rows, err := db.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Transaction
	for rows.Next() {
		var tx Transaction
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

func CheckValid (v Validatable) error {
	err := v.Validate()

	return err
}
