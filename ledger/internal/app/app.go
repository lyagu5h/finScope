package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lyagu5h/finScope/ledger/internal/service"
	"github.com/redis/go-redis/v9"

	"github.com/lyagu5h/finScope/ledger/internal/repository/pg"
)

type LedgerService = service.LedgerService
type CloseFn = func() error
func NewLedgerService(
	ctx context.Context, logger *slog.Logger,
) (service.LedgerService, func() error, error) {
	dsn := buildDSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Warn("failed to coonnect to DB: %v", slog.String("error", err.Error()))
		return nil, nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		logger.Warn("failed to ping DB: %v", slog.String("error", err.Error()))
		return nil, nil, err
	}
	
	log.Println("DB connected")

	repo := pg.New(db)

	ledgerService := service.New(
		repo.BudgetRepository,
		repo.TransactionRepository,
	)

	closeFn := func() error {
		return db.Close()
	}

	addr := getEnv("REDIS_ADDR", "localhost:6379")
	pass := os.Getenv("REDIS_PASSWORD")

	dbNum := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			dbNum = n
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       dbNum,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	log.Println("redis connected")

	return ledgerService, closeFn, nil
}

func buildDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "cashapp")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}


func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
