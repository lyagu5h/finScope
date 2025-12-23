package server

import (
	"errors"
	"strings"

	ledgerv1 "github.com/lyagu5h/finScope/ledger/internal/delivery/protos/ledger/v1"
	"github.com/lyagu5h/finScope/ledger/internal/domain"
	"github.com/lyagu5h/finScope/ledger/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapError(err error) error {
	switch {
	case errors.Is(err, service.ErrBudgetExceeded):
		return status.Error(codes.FailedPrecondition, err.Error())

	case strings.Contains(err.Error(), "validation failed"): 
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func transactionToProto(tx domain.Transaction) *ledgerv1.Transaction {
	return &ledgerv1.Transaction{
		Id:          int64(tx.ID),
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        timestamppb.New(tx.Date),
	}
}

func budgetToProto(b domain.Budget) *ledgerv1.Budget {
	return &ledgerv1.Budget{
		Category: b.Category,
		Limit:    b.Limit,
	}
}