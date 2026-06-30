package queries

import (
	"time"

	"gorm.io/gorm"
)

// AILatencyRow holds latency stats for one grouping key (a model or an
// operation) over the period. All durations are milliseconds.
type AILatencyRow struct {
	Label string  `json:"label" gorm:"column:label"`
	Calls int64   `json:"calls" gorm:"column:calls"`
	AvgMS float64 `json:"avg_ms" gorm:"column:avg_ms"`
	P50MS float64 `json:"p50_ms" gorm:"column:p50_ms"`
	P95MS float64 `json:"p95_ms" gorm:"column:p95_ms"`
	P99MS float64 `json:"p99_ms" gorm:"column:p99_ms"`
}

// AIReliabilityRow holds success/failure counts for one grouping key (a model
// or an operation) over the period.
type AIReliabilityRow struct {
	Label     string  `json:"label" gorm:"column:label"`
	Calls     int64   `json:"calls" gorm:"column:calls"`
	Successes int64   `json:"successes" gorm:"column:successes"`
	Failures  int64   `json:"failures" gorm:"column:failures"`
	ErrorRate float64 `json:"error_rate" gorm:"column:error_rate"`
}

// AIOpsMetrics is the operational view of ai_usage_logs over the last
// PeriodDays: latency percentiles, reliability and throughput. It complements
// the AI Models page (which covers spend). Drift-tolerant: if the table or a
// column is missing every aggregate simply stays zero/empty rather than
// crashing.
type AIOpsMetrics struct {
	PeriodDays   int     `json:"period_days"`
	TotalCalls   int64   `json:"total_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
	P95LatencyMS float64 `json:"p95_latency_ms"`

	LatencyByModel     []AILatencyRow `json:"latency_by_model"`
	LatencyByOperation []AILatencyRow `json:"latency_by_operation"`

	ReliabilityByOperation []AIReliabilityRow `json:"reliability_by_operation"`
	ReliabilityByModel     []AIReliabilityRow `json:"reliability_by_model"`

	CallsPerDay       []DailyCount   `json:"calls_per_day"`
	SlowestOperations []AILatencyRow `json:"slowest_operations"`
}

// GetAIOps aggregates ai_usage_logs over the last periodDays into an
// operational report (latency / reliability / throughput). It is drift-tolerant:
// if the table is missing or unreadable, or the duration_ms column hasn't been
// added yet, every aggregate stays zero/empty.
func GetAIOps(db *gorm.DB, periodDays int) (*AIOpsMetrics, error) {
	m := &AIOpsMetrics{PeriodDays: periodDays}
	since := time.Now().UTC().AddDate(0, 0, -periodDays)

	base := func() *gorm.DB {
		return db.Table("ai_usage_logs").Where("created_at >= ?", since)
	}

	// Headline scalars.
	safeScan(base().Select("COUNT(*)"), &m.TotalCalls)
	safeScan(base().Select("COALESCE(AVG(CASE WHEN success THEN 1.0 ELSE 0 END) * 100, 0)"), &m.SuccessRate)

	// Latency metrics depend on duration_ms; gate them so a missing column
	// leaves the latency aggregates empty instead of erroring on every row.
	if hasColumn(db, "ai_usage_logs", "duration_ms") {
		safeScan(base().Select("COALESCE(AVG(duration_ms), 0)"), &m.AvgLatencyMS)
		safeScan(base().Select("COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms), 0)"), &m.P95LatencyMS)

		latencyBy := func(col string, dest interface{}, limit int) {
			q := base().
				Select(`` + col + ` AS label,
					COUNT(*) AS calls,
					COALESCE(AVG(duration_ms), 0) AS avg_ms,
					COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY duration_ms), 0) AS p50_ms,
					COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms), 0) AS p95_ms,
					COALESCE(percentile_cont(0.99) WITHIN GROUP (ORDER BY duration_ms), 0) AS p99_ms`).
				Group(col).
				Order("avg_ms DESC")
			if limit > 0 {
				q = q.Limit(limit)
			}
			q.Find(dest)
		}

		latencyBy("model", &m.LatencyByModel, 0)
		latencyBy("operation", &m.LatencyByOperation, 0)
		// Slowest operations: same shape, ranked by avg latency, capped.
		latencyBy("operation", &m.SlowestOperations, 10)
	}

	// Reliability: success/failure counts + error rate %, by operation and by
	// model. These don't need duration_ms.
	reliabilityBy := func(col string, dest interface{}) {
		base().
			Select(`` + col + ` AS label,
				COUNT(*) AS calls,
				COALESCE(SUM(CASE WHEN success THEN 1 ELSE 0 END), 0) AS successes,
				COALESCE(SUM(CASE WHEN success THEN 0 ELSE 1 END), 0) AS failures,
				COALESCE(AVG(CASE WHEN success THEN 0 ELSE 1.0 END) * 100, 0) AS error_rate`).
			Group(col).
			Order("failures DESC").
			Find(dest)
	}

	reliabilityBy("operation", &m.ReliabilityByOperation)
	reliabilityBy("model", &m.ReliabilityByModel)

	// Throughput: calls per day.
	base().
		Select("DATE(created_at) AS date, COUNT(*) AS count").
		Group("DATE(created_at)").
		Order("date").
		Find(&m.CallsPerDay)

	return m, nil
}
