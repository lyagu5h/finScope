package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lyagu5h/finScope/ledger/pkg/cache"
	"github.com/lyagu5h/finScope/ledger/pkg/db"
)

type ReportSummary map[string]float64

func (s *Store) GetReportSummary(ctx context.Context, from, to time.Time) (ReportSummary, error) {
	key := fmt.Sprintf(
		"report:summary:%s:%s",
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)

	if cache.Client != nil {
		if val, err := cache.Client.Get(ctx, key).Result(); err == nil {
			var cached ReportSummary
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached, nil 
			}
		}
	}

	const q = `
		SELECT category, SUM(amount)
		FROM expenses
		WHERE date >= $1 AND date <= $2
		GROUP BY category
	`

	rows, err := db.DB.Query(q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(ReportSummary)

	for rows.Next() {
		var category string
		var sum float64
		if err := rows.Scan(&category, &sum); err != nil {
			return nil, err
		}
		result[category] = sum
	}

	if cache.Client != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = cache.Client.Set(ctx, key, data, 30*time.Second).Err()
		}
	}

	return result, nil
}

