package queries

import (
	"time"

	"gorm.io/gorm"
)

// AIModelRow is per-model usage over the period.
type AIModelRow struct {
	Model        string  `json:"model" gorm:"column:model"`
	Provider     string  `json:"provider" gorm:"column:provider"`
	Calls        int64   `json:"calls" gorm:"column:calls"`
	InputTokens  int64   `json:"input_tokens" gorm:"column:input_tokens"`
	OutputTokens int64   `json:"output_tokens" gorm:"column:output_tokens"`
	CostUSD      float64 `json:"cost_usd" gorm:"column:cost_usd"`
	AvgLatencyMS float64 `json:"avg_latency_ms" gorm:"column:avg_latency_ms"`
	Failures     int64   `json:"failures" gorm:"column:failures"`
}

// AIOperationRow is per-operation spend over the period.
type AIOperationRow struct {
	Operation string  `json:"operation" gorm:"column:operation"`
	Calls     int64   `json:"calls" gorm:"column:calls"`
	CostUSD   float64 `json:"cost_usd" gorm:"column:cost_usd"`
}

// DailyAmount is a daily USD total.
type DailyAmount struct {
	Date   time.Time `json:"date" gorm:"column:date"`
	Amount float64   `json:"amount" gorm:"column:amount"`
}

// AIModelUsage is the raw AI usage aggregated from ai_usage_logs. Counterfactual
// pricing is layered on top in the cache (it needs the rate card).
type AIModelUsage struct {
	PeriodDays        int              `json:"period_days"`
	TotalCalls        int64            `json:"total_calls"`
	TotalCostUSD      float64          `json:"total_cost_usd"`
	TotalInputTokens  int64            `json:"total_input_tokens"`
	TotalOutputTokens int64            `json:"total_output_tokens"`
	ByModel           []AIModelRow     `json:"by_model"`
	ByOperation       []AIOperationRow `json:"by_operation"`
	DailySpend        []DailyAmount    `json:"daily_spend"`
}

// GetAIModelUsage aggregates ai_usage_logs over the last periodDays. It is
// drift-tolerant: if the table is missing or unreadable (e.g. the read-only
// user lacks SELECT yet), every aggregate simply stays zero/empty.
func GetAIModelUsage(db *gorm.DB, periodDays int) (*AIModelUsage, error) {
	m := &AIModelUsage{PeriodDays: periodDays}
	since := time.Now().UTC().AddDate(0, 0, -periodDays)

	base := func() *gorm.DB {
		return db.Table("ai_usage_logs").Where("created_at >= ?", since)
	}

	safeScan(base().Select("COUNT(*)"), &m.TotalCalls)
	safeScan(base().Select("COALESCE(SUM(cost_usd), 0)"), &m.TotalCostUSD)
	safeScan(base().Select("COALESCE(SUM(input_tokens), 0)"), &m.TotalInputTokens)
	safeScan(base().Select("COALESCE(SUM(output_tokens), 0)"), &m.TotalOutputTokens)

	base().
		Select(`model, provider,
			COUNT(*) as calls,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cost_usd), 0) as cost_usd,
			COALESCE(AVG(duration_ms), 0) as avg_latency_ms,
			COALESCE(SUM(CASE WHEN success THEN 0 ELSE 1 END), 0) as failures`).
		Group("model, provider").
		Order("cost_usd DESC").
		Find(&m.ByModel)

	base().
		Select("operation, COUNT(*) as calls, COALESCE(SUM(cost_usd), 0) as cost_usd").
		Group("operation").
		Order("cost_usd DESC").
		Find(&m.ByOperation)

	base().
		Select("DATE(created_at) as date, COALESCE(SUM(cost_usd), 0) as amount").
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailySpend)

	return m, nil
}
