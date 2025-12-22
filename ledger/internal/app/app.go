package app

import (
	"context"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lyagu5h/finScope/ledger/internal/cache"
	"github.com/lyagu5h/finScope/ledger/internal/db"
	"github.com/lyagu5h/finScope/ledger/internal/service"

	"github.com/lyagu5h/finScope/ledger/internal/repository/cached"
	"github.com/lyagu5h/finScope/ledger/internal/repository/pg"
)

type CloseFn = func() error
func NewLedgerService(
	ctx context.Context, logger *slog.Logger,
) (service.LedgerService, func() error, error) {
	dbConn, err := db.InitDB(ctx, logger)

	if err != nil {
		return nil, nil, err
	}


	repo := pg.New(dbConn)

	budgetRepo := repo.BudgetRepository
	txRepo := repo.TransactionRepository

	closeFn := func() error {
		return dbConn.Close()
	}

	redisClient, err := cache.InitCache(ctx, logger)
	if err != nil {
		logger.Warn("redis disabled", slog.String("error", err.Error()))
	} else {
		budgetRepo = cached.NewBudgetRepository(
			redisClient,
			budgetRepo,
			15*time.Second,
		)
	}

	
	ledgerService := service.New(
		budgetRepo,
		txRepo,
		logger,
		redisClient,
	)

	return ledgerService, closeFn, nil
}


