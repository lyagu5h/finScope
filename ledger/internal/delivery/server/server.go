package server

import (
	"context"
	"runtime"
	"time"

	ledgerv1 "github.com/lyagu5h/finScope/ledger/internal/delivery/protos/ledger/v1"
	"github.com/lyagu5h/finScope/ledger/internal/domain"
	"github.com/lyagu5h/finScope/ledger/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	ledgerv1.UnimplementedLedgerServiceServer
	svc service.LedgerService
}

func New(svc service.LedgerService) *Server {
	return &Server{svc: svc}
}

func (s *Server) AddTransaction(
	ctx context.Context,
	req *ledgerv1.CreateTransactionRequest,
) (*ledgerv1.Transaction, error) {

	var date time.Time
	if req.Date != nil {
		date = req.Date.AsTime()
	}

	tx := domain.Transaction{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        date,
	}

	res, err := s.svc.AddTransaction(ctx, tx)
	if err != nil {
		return nil, mapError(err)
	}

	return transactionToProto(res), nil
}

func (s *Server) ListTransactions(
	ctx context.Context,
	_ *emptypb.Empty,
) (*ledgerv1.ListTransactionsResponse, error) {

	txs, err := s.svc.ListTransactions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := make([]*ledgerv1.Transaction, 0, len(txs))
	for _, tx := range txs {
		out = append(out, transactionToProto(tx))
	}

	return &ledgerv1.ListTransactionsResponse{
		Transactions: out,
	}, nil
}

func (s *Server) SetBudget(
	ctx context.Context,
	req *ledgerv1.CreateBudgetRequest,
) (*ledgerv1.Budget, error) {

	b := domain.Budget{
		Category: req.Category,
		Limit:    req.Limit,
	}

	if err := s.svc.SetBudget(ctx, b); err != nil {
		return nil, mapError(err)
	}

	return budgetToProto(b), nil
}

func (s *Server) ListBudgets(
	ctx context.Context,
	_ *emptypb.Empty,
) (*ledgerv1.ListBudgetsResponse, error) {

	budgets, err := s.svc.ListBudgets(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := make([]*ledgerv1.Budget, 0, len(budgets))
	for _, b := range budgets {
		out = append(out, budgetToProto(b))
	}

	return &ledgerv1.ListBudgetsResponse{
		Budgets: out,
	}, nil
}

func (s *Server) GetReportSummary(
	ctx context.Context,
	req *ledgerv1.ReportSummaryRequest,
) (*ledgerv1.ReportSummaryResponse, error) {

	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid from date")
	}

	to, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid to date")
	}

	summary, err := s.svc.GetReportSummary(ctx, from, to)
	if err != nil {
		return nil, mapError(err)
	}

	return &ledgerv1.ReportSummaryResponse{
		Totals: summary,
	}, nil
}

func (s *Server) BulkAddTransactions(
	ctx context.Context,
	req *ledgerv1.BulkCreateTransactionsRequest,
) (*ledgerv1.BulkCreateTransactionsResponse, error) {

	if len(req.Transactions) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"transactions list is empty",
		)
	}

	workers := int(req.Workers)
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	txs := make([]domain.Transaction, 0, len(req.Transactions))
	for _, t := range req.Transactions {
		var date time.Time
		if t.Date != nil {
			date = t.Date.AsTime()
		}
		txs = append(txs, domain.Transaction{
			Amount:      t.Amount,
			Category:    t.Category,
			Description: t.Description,
			Date:        date,
		})
	}

	result, err := s.svc.ImportTransactions(ctx, txs, workers)
	if err != nil {
		return nil, mapError(err)
	}

	errs := make([]*ledgerv1.BulkImportError, 0, len(result.Errors))
	for _, e := range result.Errors {
		errs = append(errs, &ledgerv1.BulkImportError{
			Index: uint32(e.Index),
			Error: e.Error,
		})
	}

	return &ledgerv1.BulkCreateTransactionsResponse{
		Accepted: uint32(result.Accepted),
		Rejected: uint32(result.Rejected),
		Errors:   errs,
	}, nil
}
