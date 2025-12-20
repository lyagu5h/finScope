package api

import (
	"time"
)

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Date        time.Time `json:"date"`
}

type TransactionResponse struct {
	ID          int `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Date        time.Time `json:"date"`
}

type CreateBudgetRequest struct {
	Category string `json:"category"`
	Limit float64 `json:"limit"`
}

type BudgetResponse struct {
	Category string `json:"category"`
	Limit float64 `json:"limit"`
	Period string `json:"period,omitempty"`
}



