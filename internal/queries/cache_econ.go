package queries

import (
	"time"

	"gorm.io/gorm"
)

// assumedAIExtractionUSD is a DOCUMENTED ASSUMPTION: the average cost of a
// single Haiku-tier AI recipe extraction (a few thousand tokens in/out). Every
// "saved" figure on the Cache Economics page is derived from this constant and
// is therefore an ESTIMATE, not measured spend. Tune it as the rate card moves.
const assumedAIExtractionUSD = 0.0020

// CacheEconomics quantifies what the canonical (URL-keyed extraction) and
// search caching layers save. All fields prefixed "Estimated" are ESTIMATES
// derived from assumedAIExtractionUSD, surfaced as such on the page.
//
// It is drift-tolerant: missing tables/columns leave the affected metric at
// zero/empty rather than crashing (counts no-op on a missing table; hit_count
// and is_multi_page are guarded with hasColumn).
type CacheEconomics struct {
	// Headline
	TotalCanonicalEntries   int64   `json:"total_canonical_entries"`
	FreeEntries             int64   `json:"free_entries"`    // extraction_method = json_ld (no AI)
	PaidAIEntries           int64   `json:"paid_ai_entries"` // haiku + firecrawl_haiku (AI)
	FreePct                 float64 `json:"free_pct"`        // free_entries / total
	PaidAIPct               float64 `json:"paid_ai_pct"`     // paid_ai_entries / total
	TotalSearchCacheEntries int64   `json:"total_search_cache_entries"`

	// Estimated $ saved (ESTIMATES — see AssumedAIExtractionUSD)
	AssumedAIExtractionUSD   float64 `json:"assumed_ai_extraction_usd"`
	EstimatedFreeSavingsUSD  float64 `json:"estimated_free_savings_usd"`  // free extractions that skipped AI
	EstimatedReuseSavingsUSD float64 `json:"estimated_reuse_savings_usd"` // cache hits that skipped re-extraction
	EstimatedTotalSavedUSD   float64 `json:"estimated_total_saved_usd"`
	HasHitCount              bool    `json:"has_hit_count"` // false => reuse savings unavailable, omitted
	AIReuseHits              int64   `json:"ai_reuse_hits"` // sum(hit_count) over AI-extracted entries

	// Extraction-method mix (every method present, e.g. json_ld / haiku /
	// firecrawl_json_ld / firecrawl_haiku) ordered by count.
	ExtractionMethodMix []LabelCount `json:"extraction_method_mix"`

	// Multi-page collection markers vs real cached recipes. Only populated when
	// the is_multi_page column exists (HasMultiPage gates the page section).
	HasMultiPage        bool  `json:"has_multi_page"`
	MultiPageMarkers    int64 `json:"multi_page_markers"`
	SingleRecipeEntries int64 `json:"single_recipe_entries"`

	// Cache growth: daily new canonical entries over the last 30 days.
	DailyNewCanonical []DailyCount `json:"daily_new_canonical"`
}

// GetCacheEconomics estimates the savings delivered by the canonical extraction
// cache and the search-result cache. It is drift-tolerant: a missing table or
// column simply leaves the affected metric zero/empty.
func GetCacheEconomics(db *gorm.DB) (*CacheEconomics, error) {
	m := &CacheEconomics{AssumedAIExtractionUSD: assumedAIExtractionUSD}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Headline counts.
	db.Table("canonical_recipes").Count(&m.TotalCanonicalEntries)
	db.Table("canonical_recipes").Where("extraction_method = ?", "json_ld").Count(&m.FreeEntries)
	db.Table("canonical_recipes").Where("extraction_method IN ('haiku','firecrawl_haiku')").Count(&m.PaidAIEntries)
	db.Table("search_caches").Count(&m.TotalSearchCacheEntries)

	if m.TotalCanonicalEntries > 0 {
		m.FreePct = float64(m.FreeEntries) / float64(m.TotalCanonicalEntries) * 100
		m.PaidAIPct = float64(m.PaidAIEntries) / float64(m.TotalCanonicalEntries) * 100
	}

	// Extraction-method mix (all methods, most common first).
	db.Table("canonical_recipes").
		Select("extraction_method as label, COUNT(*) as count").
		Group("extraction_method").
		Order("count DESC").
		Find(&m.ExtractionMethodMix)

	// Estimated $ saved (ESTIMATES).
	// Free json_ld extractions each avoided one paid AI extraction.
	m.EstimatedFreeSavingsUSD = float64(m.FreeEntries) * assumedAIExtractionUSD

	// Reuse savings: each cache hit on an AI-extracted entry avoided re-running
	// that paid extraction. Only meaningful when the hit_count column exists.
	if hasColumn(db, "canonical_recipes", "hit_count") {
		m.HasHitCount = true
		safeScan(db.Table("canonical_recipes").
			Where("extraction_method IN ('haiku','firecrawl_haiku')").
			Select("COALESCE(SUM(hit_count), 0)"), &m.AIReuseHits)
		m.EstimatedReuseSavingsUSD = float64(m.AIReuseHits) * assumedAIExtractionUSD
	}
	m.EstimatedTotalSavedUSD = m.EstimatedFreeSavingsUSD + m.EstimatedReuseSavingsUSD

	// Multi-page collection markers vs real cached recipes (column-guarded).
	if hasColumn(db, "canonical_recipes", "is_multi_page") {
		m.HasMultiPage = true
		db.Table("canonical_recipes").Where("is_multi_page = ?", true).Count(&m.MultiPageMarkers)
		db.Table("canonical_recipes").Where("is_multi_page = ?", false).Count(&m.SingleRecipeEntries)
	} else {
		m.SingleRecipeEntries = m.TotalCanonicalEntries
	}

	// Cache growth: daily new canonical entries (last 30 days).
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	db.Table("canonical_recipes").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailyNewCanonical)

	return m, nil
}
