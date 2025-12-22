package ledger

import (
	"context"
	"log/slog"

	"github.com/lyagu5h/finScope/ledger/internal/app"
	"github.com/lyagu5h/finScope/ledger/internal/domain"
	"github.com/lyagu5h/finScope/ledger/internal/service"
)

type (
	LedgerService = service.LedgerService
	Budget = domain.Budget
	Transaction = domain.Transaction
	Validatable = domain.Validatable
)

var ErrBudgetExceeded = service.ErrBudgetExceeded

func CheckValid(v Validatable) error {
	return domain.CheckValid(v)
}

func NewLedgerService(ctx context.Context, logger *slog.Logger) (LedgerService, app.CloseFn,error) {
	return app.NewLedgerService(ctx, logger)
}


