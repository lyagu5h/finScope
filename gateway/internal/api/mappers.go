package api

import (
	ledgerv1 "github.com/lyagu5h/finScope/gateway/internal/delivery/protos/ledger/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoCreateTransaction(req CreateTransactionRequest) *ledgerv1.CreateTransactionRequest {
	var ts *timestamppb.Timestamp
	if !req.Date.IsZero() {
		ts = timestamppb.New(req.Date)
	}
	
	return &ledgerv1.CreateTransactionRequest{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        ts,
	}
}

func toTransactionDTOFromProto(tx *ledgerv1.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.Id,
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date.AsTime(),
	}
}

func toProtoCreateBudget(req CreateBudgetRequest) *ledgerv1.CreateBudgetRequest {
	return &ledgerv1.CreateBudgetRequest{
		Category: req.Category,
		Limit:    req.Limit,
	}
}

func toBudgetDTOFromProto(b *ledgerv1.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
	}
}


// func toReportSummaryDTO(totals map[string]float64) map[string]float64 {
// 	return totals
// }

func toProtoBulkCreateTransactions(req BulkCreateTransactionsRequest) *ledgerv1.BulkCreateTransactionsRequest {
	out := make([]*ledgerv1.CreateTransactionRequest, 0, len(req.Transactions))

	for _, tx := range req.Transactions {
		out = append(out, toProtoCreateTransaction(tx))
	}

	return &ledgerv1.BulkCreateTransactionsRequest{
		Transactions: out,
	}
}

func toBulkResponseDTO(res *ledgerv1.BulkCreateTransactionsResponse) BulkCreateTransactionsResponse {
	errors := make([]BulkImportErrorResponse, 0, len(res.Errors))

	for _, e := range res.Errors {
		errors = append(errors, BulkImportErrorResponse{
			Index: int(e.Index),
			Error: e.Error,
		})
	}

	return BulkCreateTransactionsResponse{
		Accepted: int(res.Accepted),
		Rejected: int(res.Rejected),
		Errors:   errors,
	}
}