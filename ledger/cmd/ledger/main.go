package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/lyagu5h/finScope/ledger/internal/app"
	ledgerv1 "github.com/lyagu5h/finScope/ledger/internal/delivery/protos/ledger/v1"
	"github.com/lyagu5h/finScope/ledger/internal/delivery/server"
	"google.golang.org/grpc"
)


func main() {
	ctx := context.Background()
	handlerLog := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handlerLog)
	
	svc, closeFn, err := app.NewLedgerService(ctx, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFn()

	grpcServer := grpc.NewServer()
	ledgerGrpcServer := server.New(svc)

	ledgerv1.RegisterLedgerServiceServer(
		grpcServer,
		ledgerGrpcServer,
	)

	port := os.Getenv("LEDGER_GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("gRPC server started", slog.String("port", port))
	grpcServer.Serve(lis)
}