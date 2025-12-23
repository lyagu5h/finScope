package domain

import (
	"errors"
	"time"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

func (tx Transaction) Validate() error {

	if tx.Amount == 0 || tx.Amount < 0 {
		return errors.New("validation failed: transaction amount should be > 0")
	}

	if tx.Category == "" {
		return errors.New("validation failed: budget category cannot be empty")
	}

	return nil
}
