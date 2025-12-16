package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

type Budget struct {
	Category string `json:"category"`
	Limit float64 `json:"limit"`
	Period string `json:"period,omitempty"`
}

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

type Validatable interface {
	Validate() error
}

type Store struct {
	mu  sync.RWMutex
	txs []Transaction
	budgets map[string]Budget	
}

func NewStore() *Store {
	return &Store{
		budgets: make(map[string]Budget),
		txs: []Transaction{},
	}
}

func (s *Store) SetBudget(b Budget) error {
	if b.Category == "" {
		return errors.New("budget's category cannot be empty")
	}

	if b.Limit <= 0 {
		return errors.New("budget's limit cannot be <= 0")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.budgets[b.Category] = b
	return nil
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

	err := tx.Validate()

	if err != nil {
		return fmt.Errorf("AddTransaction: %s", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if b, ok := s.budgets[tx.Category]; ok {
		current := 0.0
		for _, t := range s.txs {
			if t.Category == tx.Category {
				current += t.Amount
			}
		}

		if current + tx.Amount > b.Limit {
			return errors.New("budget exceeded")
		}
	}

	tx.ID = len(s.txs) + 1
	if tx.Date.IsZero() {
		tx.Date = time.Now()
	}
	s.txs = append(s.txs, tx)
	return nil
}

func (s *Store) ListTransactions() []Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tmp := make([]Transaction, len(s.txs))
	copy(tmp, s.txs)
	return tmp
}

func (tx Transaction) Validate() error {
	
	if tx.Amount == 0 || tx.Amount < 0 {
		return errors.New("amount of transaction cannot be = 0 or < 0")
	}

	if tx.Category == "" {
		return errors.New("category cannot be empty")
	}

	if tx.Date.IsZero() {
		return errors.New("date must be set")
	}

	return nil
}