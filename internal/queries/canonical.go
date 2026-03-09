package queries

import (
	"time"

	"gorm.io/gorm"
)

type CanonicalMetrics struct {
	TotalEntries             int64       `json:"total_entries"`
	NewToday                 int64       `json:"new_today"`
	AvgHitCount              float64     `json:"avg_hit_count"`
	ZeroHitEntries           int64       `json:"zero_hit_entries"`
	ExtractionMethodDist     []LabelCount `json:"extraction_method_distribution"`
	TopByHits                []TopEntry   `json:"top_by_hits"`
	EntriesApproachingTTL    int64       `json:"entries_approaching_ttl"`
	HotEntriesEligible       int64       `json:"hot_entries_eligible"`
	StaleEntries90d          int64       `json:"stale_entries_90d"`
	FirecrawlCallsThisMonth  int64       `json:"firecrawl_calls_this_month"`
	FirecrawlTotalCalls      int64       `json:"firecrawl_total_calls"`
	TotalAIExtractionsSaved  int64       `json:"total_ai_extractions_saved"`
	DailyNew                 []DailyCount `json:"daily_new"`
	TopDomains               []TopEntry   `json:"top_domains"`
}

func GetCanonicalMetrics(db *gorm.DB) (*CanonicalMetrics, error) {
	m := &CanonicalMetrics{}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	db.Table("canonical_recipes").Count(&m.TotalEntries)
	db.Table("canonical_recipes").Where("created_at >= ?", today).Count(&m.NewToday)
	db.Table("canonical_recipes").Where("hit_count = 0").Count(&m.ZeroHitEntries)

	safeScan(db.Table("canonical_recipes").Select("COALESCE(AVG(hit_count), 0)"), &m.AvgHitCount)

	// Extraction method distribution
	db.Table("canonical_recipes").
		Select("extraction_method as label, COUNT(*) as count").
		Group("extraction_method").
		Find(&m.ExtractionMethodDist)

	// Top by hits
	db.Table("canonical_recipes").
		Select("original_url as label, hit_count as count, extraction_method as extra").
		Order("hit_count DESC").
		Limit(25).
		Find(&m.TopByHits)

	// Entries approaching TTL (within 1 day of 7-day TTL)
	sixDaysAgo := now.Add(-6 * 24 * time.Hour)
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	db.Table("canonical_recipes").
		Where("fetched_at BETWEEN ? AND ?", sevenDaysAgo, sixDaysAgo).
		Count(&m.EntriesApproachingTTL)

	// Hot entries eligible for refresh
	db.Table("canonical_recipes").
		Where("hit_count >= 5 AND fetched_at BETWEEN ? AND ?", sevenDaysAgo, sixDaysAgo).
		Count(&m.HotEntriesEligible)

	// Stale entries (>90d no access)
	ninetyDaysAgo := now.Add(-90 * 24 * time.Hour)
	db.Table("canonical_recipes").
		Where("last_accessed_at < ?", ninetyDaysAgo).
		Count(&m.StaleEntries90d)

	// Firecrawl calls this month
	db.Table("canonical_recipes").
		Where("created_at >= ? AND extraction_method IN ('firecrawl_json_ld','firecrawl_haiku')", monthStart).
		Count(&m.FirecrawlCallsThisMonth)

	// Firecrawl total calls
	db.Table("canonical_recipes").
		Where("extraction_method IN ('firecrawl_json_ld','firecrawl_haiku')").
		Count(&m.FirecrawlTotalCalls)

	// AI extractions saved: sum of hit_count for haiku-extracted entries
	safeScan(db.Table("canonical_recipes").
		Where("extraction_method IN ('haiku','firecrawl_haiku')").
		Select("COALESCE(SUM(hit_count), 0)"), &m.TotalAIExtractionsSaved)

	// Daily new (30 days)
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	db.Table("canonical_recipes").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailyNew)

	// Top source domains
	db.Raw(`SELECT
		SUBSTRING(original_url FROM '://([^/]+)') as label,
		COUNT(*) as count
		FROM canonical_recipes
		WHERE original_url != ''
		GROUP BY label
		ORDER BY count DESC
		LIMIT 20`).Find(&m.TopDomains)

	return m, nil
}
