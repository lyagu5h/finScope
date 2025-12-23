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
	ID          int64 `json:"id"`
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
}

type BulkCreateTransactionsRequest struct {
	Transactions []CreateTransactionRequest `json:"transactions"`
	Workers      int                         `json:"workers,omitempty"`
}

type BulkImportErrorResponse struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type BulkCreateTransactionsResponse struct {
	Accepted int                      `json:"accepted"`
	Rejected int                      `json:"rejected"`
	Errors   []BulkImportErrorResponse `json:"errors"`
}



