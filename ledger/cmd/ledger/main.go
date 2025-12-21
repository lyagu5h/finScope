package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lyagu5h/finScope/ledger/pkg/ledger"
)

func main() {
	ctx := context.Background()
	handlerLog := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handlerLog)
	
	_, closeFn, err := ledger.NewLedgerService(ctx, logger)

	if err != nil {
		logger.Error("failed to fabric ledger service", slog.String("error", err.Error()))
	}
	defer closeFn()

	logger.Info("ledger is up")
}
