package queries

import (
	"time"

	"gorm.io/gorm"
)

// defaultVideoDailyBudgetUSD mirrors the API's VIDEO_IMPORT_DAILY_BUDGET_USD
// default ($25/day kill switch). The dashboard reads the DB read-only and has no
// access to the API's env, so it surfaces the documented default for context.
const defaultVideoDailyBudgetUSD = 25.0

// VideoPlatformRow is per-platform economics for premium video imports.
type VideoPlatformRow struct {
	Platform  string  `json:"platform" gorm:"column:platform"`
	Imports   int64   `json:"imports" gorm:"column:imports"`
	CostUSD   float64 `json:"cost_usd" gorm:"column:cost_usd"`
	CacheHits int64   `json:"cache_hits" gorm:"column:cache_hits"`
	Failures  int64   `json:"failures" gorm:"column:failures"`
}

// VideoDailyRow is a single day's video-import spend and volume.
type VideoDailyRow struct {
	Date    time.Time `json:"date" gorm:"column:date"`
	Spend   float64   `json:"spend" gorm:"column:spend"`
	Imports int64     `json:"imports" gorm:"column:imports"`
}

// VideoEconomics is the economics of the premium video->recipe feature over the
// last periodDays, aggregated from the video_imports table.
type VideoEconomics struct {
	PeriodDays     int                `json:"period_days"`
	TotalImports   int64              `json:"total_imports"`
	TotalSpendUSD  float64            `json:"total_spend_usd"`
	PaidImports    int64              `json:"paid_imports"`
	SuccessCount   int64              `json:"success_count"`
	FailedCount    int64              `json:"failed_count"`
	SuccessRate    float64            `json:"success_rate"` // percent 0-100
	CacheHits      int64              `json:"cache_hits"`
	CacheHitRate   float64            `json:"cache_hit_rate"` // percent 0-100
	AvgCostPerPaid float64            `json:"avg_cost_per_paid_import"`
	TodaySpendUSD  float64            `json:"today_spend_usd"`
	DailyBudgetUSD float64            `json:"daily_budget_usd"`
	ByPlatform     []VideoPlatformRow `json:"by_platform"`
	ByStatus       []LabelCount       `json:"by_status"`
	Daily          []VideoDailyRow    `json:"daily"`
}

// GetVideoEconomics aggregates video_imports over the last periodDays. It is
// drift-tolerant: the table only exists once premium video import is enabled
// (a provider key is set), and individual columns are guarded with hasColumn.
// A missing table or column simply leaves the affected metric zero/empty rather
// than crashing, so the page falls back to a clean empty state.
func GetVideoEconomics(db *gorm.DB, periodDays int) (*VideoEconomics, error) {
	m := &VideoEconomics{PeriodDays: periodDays, DailyBudgetUSD: defaultVideoDailyBudgetUSD}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	since := today.AddDate(0, 0, -periodDays)

	const table = "video_imports"
	hasDeleted := hasColumn(db, table, "deleted_at")
	hasCost := hasColumn(db, table, "cost_usd")
	hasStatus := hasColumn(db, table, "status")
	hasPlatform := hasColumn(db, table, "platform")
	hasCacheHit := hasColumn(db, table, "cache_hit")

	base := func() *gorm.DB {
		q := db.Table(table).Where("created_at >= ?", since)
		if hasDeleted {
			q = q.Where("deleted_at IS NULL")
		}
		return q
	}

	// Conditional SQL fragments so a missing column degrades to a literal 0
	// instead of erroring the whole grouped query.
	costExpr := "0"
	if hasCost {
		costExpr = "COALESCE(SUM(cost_usd), 0)"
	}
	cacheExpr := "0"
	if hasCacheHit {
		cacheExpr = "COALESCE(SUM(CASE WHEN cache_hit THEN 1 ELSE 0 END), 0)"
	} else if hasCost {
		// Infer cache hits from zero metered cost when no cache_hit column exists.
		cacheExpr = "COALESCE(SUM(CASE WHEN cost_usd = 0 THEN 1 ELSE 0 END), 0)"
	}
	failExpr := "0"
	if hasStatus {
		failExpr = "COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0)"
	}

	// Headline scalars.
	safeScan(base().Select("COUNT(*)"), &m.TotalImports)

	if hasCost {
		safeScan(base().Select("COALESCE(SUM(cost_usd), 0)"), &m.TotalSpendUSD)
		safeScan(base().Where("cost_usd > ?", 0).Select("COUNT(*)"), &m.PaidImports)
		safeScan(base().Where("created_at >= ?", today).Select("COALESCE(SUM(cost_usd), 0)"), &m.TodaySpendUSD)
	}

	if hasStatus {
		safeScan(base().Where("status = ?", "done").Select("COUNT(*)"), &m.SuccessCount)
		safeScan(base().Where("status = ?", "failed").Select("COUNT(*)"), &m.FailedCount)
	}

	if hasCacheHit {
		safeScan(base().Where("cache_hit = ?", true).Select("COUNT(*)"), &m.CacheHits)
	} else if hasCost {
		safeScan(base().Where("cost_usd = ?", 0).Select("COUNT(*)"), &m.CacheHits)
	}

	if m.TotalImports > 0 {
		m.SuccessRate = float64(m.SuccessCount) / float64(m.TotalImports) * 100
		m.CacheHitRate = float64(m.CacheHits) / float64(m.TotalImports) * 100
	}
	if m.PaidImports > 0 {
		m.AvgCostPerPaid = m.TotalSpendUSD / float64(m.PaidImports)
	}

	// Per-platform economics.
	if hasPlatform {
		base().
			Select("platform, COUNT(*) as imports, " +
				costExpr + " as cost_usd, " +
				cacheExpr + " as cache_hits, " +
				failExpr + " as failures").
			Group("platform").
			Order("imports DESC").
			Find(&m.ByPlatform)
	}

	// Success/failure (and any other lifecycle state) breakdown.
	if hasStatus {
		base().
			Select("status as label, COUNT(*) as count").
			Group("status").
			Order("count DESC").
			Find(&m.ByStatus)
	}

	// Daily spend + volume over the period.
	base().
		Select("DATE(created_at) as date, " + costExpr + " as spend, COUNT(*) as imports").
		Group("DATE(created_at)").
		Order("date").
		Find(&m.Daily)

	return m, nil
}
