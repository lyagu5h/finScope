package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
	"github.com/redis/go-redis/v9"
)

var ErrBudgetExceeded = errors.New("budget exceeded")
type BulkImportResult struct {
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
	Errors   []BulkImportError `json:"errors"`
}

type BulkImportError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type LedgerService interface {
	SetBudget(ctx context.Context, b domain.Budget) error
	ListBudgets(ctx context.Context) ([]domain.Budget, error)

	AddTransaction(ctx context.Context, t domain.Transaction) (domain.Transaction, error) 
	ListTransactions(ctx context.Context) ([]domain.Transaction, error)

	GetReportSummary(ctx context.Context, from, to time.Time) (map[string]float64, error)
	ImportTransactions(ctx context.Context, txs []domain.Transaction, workers int,) (BulkImportResult, error)
}

type ledger struct {
	budgets domain.BudgetRepository
	transactions domain.TransactionRepository
	log *slog.Logger
	cache *redis.Client
}

type importJob struct {
	Index int
	Tx    domain.Transaction
}

type importResult struct {
	Index int
	Err   error
}


func New(
	budgetsRepo domain.BudgetRepository,
	transactionsRepo domain.TransactionRepository,
	logger *slog.Logger,
	redisClient *redis.Client,
) LedgerService {
	return &ledger{
		budgets:      budgetsRepo,
		transactions: transactionsRepo,
		log: logger,
		cache: redisClient,
	}
}

func (svc *ledger) AddTransaction(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	svc.log.Info(
		"transaction add requested",
		slog.String("category", t.Category),
		slog.Float64("amount", t.Amount),
	)

	if t.Date.IsZero() {
		t.Date = time.Now()
	}
	if err := t.Validate(); err != nil {
		return t, err
	}

	budget, ok, err := svc.budgets.GetByCategory(ctx, t.Category)
	if err != nil {
		return t, err
	}

	if ok {
		current, err := svc.transactions.SumByCategory(ctx, t.Category)
		if err != nil {
			return t, err
		}

		if current+t.Amount > budget.Limit {
			svc.log.Info(
				"budget exceeded",
				slog.String("error", ErrBudgetExceeded.Error()),
			)

			return t, ErrBudgetExceeded
		}	
	}

	if err := svc.transactions.Add(ctx, &t); err != nil {
		return t, err
	}

	return t, nil
}

func (svc *ledger) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return svc.transactions.List(ctx)
}

func (svc *ledger) SetBudget(ctx context.Context, b domain.Budget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	svc.log.Info(
		"budget set",
		slog.String("category", b.Category),
		slog.Float64("limit", b.Limit),
	)
	return svc.budgets.Upsert(ctx, b)
}

func (svc *ledger) ListBudgets(ctx context.Context) ([]domain.Budget, error) {
	return svc.budgets.List(ctx)
}

func (svc *ledger) GetReportSummary(
	ctx context.Context,
	from, to time.Time,
) (map[string]float64, error) {
	svc.log.Info(
		"report requested",
		slog.String("from", from.Format("2006-01-02")),
		slog.String("to", to.Format("2006-01-02")),
	)

	cacheKey := "report:summary:" +
		from.Format("2006-01-02") + ":" +
		to.Format("2006-01-02")

	if svc.cache != nil {
		if data, err := svc.cache.Get(ctx, cacheKey).Bytes(); err == nil {
			var cached map[string]float64
			if err := json.Unmarshal(data, &cached); err == nil {
				svc.log.Info("report cache hit", slog.String("key", cacheKey))
				return cached, nil
			}
		}
	}

	svc.log.Info("report cache miss", slog.String("key", cacheKey))

	categories, err := svc.transactions.ListCategories(ctx)
	if err != nil {
		svc.log.Error("failed to list categories", slog.String("error", err.Error()))
		return nil, err
	}

	result := make(map[string]float64)

	type item struct {
		category string
		sum      float64
	}

	resultsCh := make(chan item, len(categories))
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	var once sync.Once

	ticker := time.NewTicker(400 * time.Millisecond)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				svc.log.Debug(
					"report in progress",
					slog.String("from", from.Format("2006-01-02")),
					slog.String("to", to.Format("2006-01-02")),
				)
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	for _, category := range categories {
		wg.Add(1)

		go func(cat string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			sum, err := svc.transactions.SumByCategoryAndPeriod(
				ctx,
				cat,
				from,
				to,
			)
			if err != nil {
				once.Do(func() {
					errCh <- err
				})
				return
			}

			if sum > 0 {
				select {
				case resultsCh <- item{category: cat, sum: sum}:
				case <-ctx.Done():
					return
				}
			}
		}(category)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			svc.log.Warn(
				"report cancelled",
				slog.String("reason", ctx.Err().Error()),
			)
			return nil, ctx.Err()

		case err := <-errCh:
			if err != nil {
				svc.log.Error(
					"report failed",
					slog.String("error", err.Error()),
				)
				return nil, err
			}

		case it, ok := <-resultsCh:
			if !ok {
				if svc.cache != nil {
					if data, err := json.Marshal(result); err == nil {
						_ = svc.cache.Set(ctx, cacheKey, data, 30*time.Second).Err()
						svc.log.Info("report cached", slog.String("key", cacheKey))
					}
				}

				svc.log.Info(
					"report completed",
					slog.Int("categories", len(result)),
					slog.String("from", from.Format("2006-01-02")),
					slog.String("to", to.Format("2006-01-02")),
				)

				return result, nil
			}

			result[it.category] = it.sum
		}
	}
}


func (svc *ledger) ImportTransactions(
	ctx context.Context,
	txs []domain.Transaction,
	workers int,
) (BulkImportResult, error) {

	jobs := make(chan importJob)
	results := make(chan importResult)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}

					_, err := svc.AddTransaction(ctx, job.Tx)
					results <- importResult{
						Index: job.Index,
						Err:   err,
					}
				}
			}
		}(i)
	}

	go func() {
		for i, tx := range txs {
			select {
			case <-ctx.Done():
				return
			case jobs <- importJob{Index: i, Tx: tx}:
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var accepted int64
	var rejected int64

	summary := BulkImportResult{
		Errors: make([]BulkImportError, 0),
	}

	for res := range results {
		if res.Err == nil {
			atomic.AddInt64(&accepted, 1)
			continue
		}

		atomic.AddInt64(&rejected, 1)
		summary.Errors = append(summary.Errors, BulkImportError{
			Index: res.Index,
			Error: res.Err.Error(),
		})
	}

	summary.Accepted = int(accepted)
	summary.Rejected = int(rejected)

	if ctx.Err() != nil {
		return summary, ctx.Err()
	}

	return summary, nil
}