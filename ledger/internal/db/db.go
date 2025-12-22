package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
)

func InitDB(ctx context.Context, logger *slog.Logger) (*sql.DB, error){
	dsn := buildDSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Warn("failed to coonnect to DB", slog.String("error", err.Error()))
		return nil,err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.PingContext(ctx); err != nil {
		logger.Warn("failed to ping DB: %v", slog.String("error", err.Error()))
		return nil, err
	}

	logger.Info("DB connected", slog.String("dsn", dsn))

	return db, nil
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