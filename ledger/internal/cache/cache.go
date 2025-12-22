package cache

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func InitCache(ctx context.Context, logger *slog.Logger) (*redis.Client, error) {
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

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger.Info("redis connected", slog.String("addr", addr), slog.Int("db_num", dbNum))

	return rdb, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
