package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Transaction struct {
	ID int
	Amount float64
	Category string
	Description string
	Date time.Time
}

type Store struct {
	mu sync.RWMutex
	txs []Transaction
}

func NewStore() *Store {
	return &Store{
		txs: []Transaction{},
	}
}

func (s *Store) AddTransaction(tx Transaction) error {
	
	if tx.Amount == 0 || tx.Amount < 0 {
		return errors.New("amount of transaction cannot be == 0 or < 0")
	}

	if tx.Category == "" {
		return errors.New("category cannot be empty")
	}

	s.mu.Lock()
	s.txs = append(s.txs, tx)
	s.mu.Unlock()
	return nil
}

func (s *Store) ListTransactions() []Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tmp := make([]Transaction, len(s.txs))
	copy(tmp, s.txs)
	return tmp
} 

func main() {
	store := NewStore()

	transactions := []Transaction{
		{
			Amount: 1499.90, 
			Category: "Food", 
			Description: "Покупка продуктов в Магнит", 
			Date: time.Date(2025, 1, 12, 14, 30, 0, 0, time.UTC),
		},
		{
			Amount: 56000.00, 
			Category: "Salary", 
			Description: "Зарплата", 
			Date: time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC),
		},
	}

	for _, t := range transactions {
		err := store.AddTransaction(t)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println(store.ListTransactions())
}