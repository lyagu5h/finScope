package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/lyagu5h/finScope/gateway/internal/api"
	"github.com/lyagu5h/finScope/ledger/pkg/ledger"
)

func main() {
	mux := http.NewServeMux()
	handlerLog := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handlerLog)
	ledgerSvc, closeFn, err := ledger.NewLedgerService(context.Background(), logger)

	if err != nil {
		logger.Error("failed to fabric ledger service", slog.String("error", err.Error()))
	}
	defer closeFn()
	handler := api.NewHandler(ledgerSvc, logger)

	port := ":8080"

	handler.RegisterRoutes(mux)
	logger.Info("API Gateway starting on port :8080", slog.String("port", port))
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}