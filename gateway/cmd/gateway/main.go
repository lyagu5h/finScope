package main

import (
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
	store := ledger.NewStore(logger)
	handler := api.NewHandler(store)

	handler.RegisterRoutes(mux)
	logger.Info("API Gateway starting on port :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}