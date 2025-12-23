package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lyagu5h/finScope/gateway/internal/api"
	"github.com/lyagu5h/finScope/gateway/internal/delivery/client"
)

// @title FinScope API
// @version 1.0
// @description HTTP API Gateway for FinScope Ledger
// @host localhost:8080
// @BasePath /

func main() {
	handlerLog := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handlerLog)

	ledgerAddr := os.Getenv("LEDGER_ADDR")
	if ledgerAddr == "" {
		ledgerAddr = "localhost:50051"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	timeout := 2 * time.Second

	ledgerClient, err := client.New(ledgerAddr)
	if err != nil {
		logger.Error("failed to create grpc client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler := api.NewHandler(ledgerClient, logger, timeout)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Info("API Gateway starting", slog.String("addr", httpAddr))
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
