package api

import "github.com/lyagu5h/finScope/ledger/pkg/ledger"

func toTransactionLedger(r CreateTransactionRequest) ledger.Transaction {
	return ledger.Transaction{
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		Date:        r.Date,
	}
}

func toTransactionDTO(tx ledger.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.ID,
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date,
	}
}

func toBudgetLedger(b CreateBudgetRequest) ledger.Budget {
	return ledger.Budget{
		Category: b.Category,
		Limit:    b.Limit,
	}
}

func toBudgetDTO(b ledger.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
		Period:   b.Period,
	}
}