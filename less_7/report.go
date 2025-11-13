package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/you/monorepo/ledger/internal/cache"
	"github.com/you/monorepo/ledger/internal/db"
)

type CategorySummary struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

func GetReportSummary(ctx context.Context, from, to time.Time) ([]CategorySummary, error) {
	key := fmt.Sprintf("report:summary:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	if val, err := cache.Client.Get(ctx, key).Result(); err == nil {
		var cached []CategorySummary
		_ = json.Unmarshal([]byte(val), &cached)
		return cached, nil
	}

	const q = `
	SELECT category, COALESCE(SUM(amount),0) AS total
	FROM expenses
	WHERE date BETWEEN $1 AND $2
	GROUP BY category`
	rows, err := db.DB.QueryContext(ctx, q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []CategorySummary
	for rows.Next() {
		var s CategorySummary
		rows.Scan(&s.Category, &s.Total)
		result = append(result, s)
	}
	data, _ := json.Marshal(result)
	cache.Client.Set(ctx, key, data, 30*time.Second)
	return result, nil
}
