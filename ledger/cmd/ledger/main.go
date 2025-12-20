package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lyagu5h/finScope/ledger/pkg/cache"
	"github.com/lyagu5h/finScope/ledger/pkg/db"
	"github.com/lyagu5h/finScope/ledger/pkg/ledger"
)

func main() {
	ctx := context.Background()
	handlerLog := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	log := slog.New(handlerLog)
	store := ledger.NewStore(log)

	db.InitDB()
	cache.InitRedis()

	_ = store.SetBudget(ctx, ledger.Budget{Category: "еда", Limit: 5000})
	_ = store.SetBudget(ctx, ledger.Budget{Category: "транспорт", Limit: 1500})

	tx := ledger.Transaction{
		Amount:      1200,
		Category:    "еда",
		Description: "Продукты",
		Date:        time.Now(),
	}

	if err := ledger.CheckValid(tx); err != nil {
		log.Error("CheckValid failed", slog.String("err", err.Error()))
	} else if err := store.AddTransaction(tx); err != nil {
		log.Error("AddTransaction A failed", slog.String("err", err.Error()))
	} else {
		log.Info("AddTransaction A OK")
	}

	if err := store.AddTransaction(ledger.Transaction{
		Amount:      999999,
		Category:    "еда",
		Description: "Неадекватная трата",
		Date:        time.Now(),
	}); err != nil {
		log.Warn("AddTransaction B expected error", slog.String("err", err.Error()))
	}

	txs, _ := store.ListTransactions()
	for _, t := range txs {
		log.Info("transaction",
			slog.Int("id", t.ID),
			slog.String("category", t.Category),
			slog.Float64("amount", t.Amount),
			slog.String("desc", t.Description),
			slog.Time("date", t.Date),
		)
	}

	bs, _ := store.ListBudgets(ctx)
	for _, b := range bs {
		log.Info("budget",
			slog.String("category", b.Category),
			slog.Float64("limit", b.Limit),
		)
	}

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	report, err := store.GetReportSummary(context.Background(), from, to)
	if err != nil {
		log.Error("GetReportSummary failed", slog.String("error", err.Error()))
	} else {
		for category, sum := range report {
			log.Info("report item",
				slog.String("category", category),
				slog.Float64("sum", sum),
			)
		}
	}
}
