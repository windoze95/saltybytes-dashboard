package queries

import (
	"time"

	"gorm.io/gorm"
)

type SearchCacheMetrics struct {
	TotalCachedQueries    int64       `json:"total_cached_queries"`
	AvgHitCount           float64     `json:"avg_hit_count"`
	ZeroHitEntries        int64       `json:"zero_hit_entries"`
	EntriesWithEmbeddings int64       `json:"entries_with_embeddings"`
	EmbeddingCoverage     float64     `json:"embedding_coverage"`
	AvgResultsPerQuery    float64     `json:"avg_results_per_query"`
	TopQueries            []TopEntry  `json:"top_queries"`
	EntriesApproachingTTL int64       `json:"entries_approaching_ttl"`
	StaleEntries          int64       `json:"stale_entries"`
	HotQueriesEligible    int64       `json:"hot_queries_eligible"`
	StaleEntries30d       int64       `json:"stale_entries_30d"`
	DailyVolume           []DailyCount `json:"daily_volume"`
}

func GetSearchCacheMetrics(db *gorm.DB) (*SearchCacheMetrics, error) {
	m := &SearchCacheMetrics{}
	now := time.Now().UTC()

	db.Table("search_caches").Count(&m.TotalCachedQueries)
	db.Table("search_caches").Where("hit_count = 0").Count(&m.ZeroHitEntries)
	db.Table("search_caches").Where("embedding IS NOT NULL").Count(&m.EntriesWithEmbeddings)

	if m.TotalCachedQueries > 0 {
		m.EmbeddingCoverage = float64(m.EntriesWithEmbeddings) / float64(m.TotalCachedQueries) * 100
	}

	db.Table("search_caches").
		Select("COALESCE(AVG(hit_count), 0)").
		Row().Scan(&m.AvgHitCount)

	db.Table("search_caches").
		Select("COALESCE(AVG(result_count), 0)").
		Row().Scan(&m.AvgResultsPerQuery)

	// Top queries by hits
	db.Table("search_caches").
		Select("normalized_query as label, hit_count as count").
		Order("hit_count DESC").
		Limit(25).
		Find(&m.TopQueries)

	// Entries approaching TTL (22-24h old)
	ttl22h := now.Add(-22 * time.Hour)
	ttl24h := now.Add(-24 * time.Hour)
	db.Table("search_caches").
		Where("fetched_at BETWEEN ? AND ?", ttl24h, ttl22h).
		Count(&m.EntriesApproachingTTL)

	// Stale entries (past 24h TTL)
	db.Table("search_caches").
		Where("fetched_at < ?", ttl24h).
		Count(&m.StaleEntries)

	// Hot queries eligible for refresh (hit_count >= 10, within TTL, approaching refresh window)
	staleAt := now.Add(-24 * time.Hour)
	refreshAt := now.Add(-22 * time.Hour)
	db.Table("search_caches").
		Where("hit_count >= 10 AND fetched_at > ? AND fetched_at < ?", staleAt, refreshAt).
		Count(&m.HotQueriesEligible)

	// Stale entries pending cleanup (30 days)
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
	db.Table("search_caches").
		Where("last_accessed_at < ?", thirtyDaysAgo).
		Count(&m.StaleEntries30d)

	// Daily new cache entries (30 days)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	daysAgo30 := today.AddDate(0, 0, -30)
	db.Table("search_caches").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", daysAgo30).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailyVolume)

	return m, nil
}
