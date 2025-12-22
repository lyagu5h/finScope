package cached

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lyagu5h/finScope/ledger/internal/domain"
	"github.com/redis/go-redis/v9"
)

const budgetsCacheKey = "budgets:all"

type BudgetRepository struct {
	cache *redis.Client
	next domain.BudgetRepository
	ttl time.Duration
}

func NewBudgetRepository(cache *redis.Client, next domain.BudgetRepository, ttl time.Duration) domain.BudgetRepository {
	return &BudgetRepository{cache: cache, next: next, ttl: ttl}
}

func (r *BudgetRepository) List(ctx context.Context) ([]domain.Budget, error) {
	if data, err := r.cache.Get(ctx, budgetsCacheKey).Bytes(); err == nil {
		var cached []domain.Budget
		if err := json.Unmarshal(data, &cached); err == nil {
			return cached, nil
		}
	}

	budgets, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(budgets); err == nil {
		_ = r.cache.Set(ctx, budgetsCacheKey, data, r.ttl).Err()
	}

	return budgets, nil
}

func (r *BudgetRepository) Upsert(ctx context.Context, b domain.Budget) error {
	if err := r.next.Upsert(ctx, b); err != nil {
		return err
	}

	_ = r.cache.Del(ctx, budgetsCacheKey).Err()
	return nil
}

func (r *BudgetRepository) GetByCategory(
	ctx context.Context,
	category string,
) (domain.Budget, bool, error) {
	return r.next.GetByCategory(ctx, category)
}