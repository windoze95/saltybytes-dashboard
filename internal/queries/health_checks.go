package queries

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func GetHealthChecks(db *gorm.DB) []HealthCheck {
	checks := []HealthCheck{}
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)

	// Recipes without trees
	var count int64
	db.Raw("SELECT COUNT(*) FROM recipes WHERE tree_id IS NULL AND deleted_at IS NULL AND created_at < ?", oneHourAgo).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Recipes without trees",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d recipes missing tree_id", count),
	})

	// Trees without root nodes
	db.Raw("SELECT COUNT(*) FROM recipe_trees WHERE root_node_id IS NULL").Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Trees without root nodes",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d trees missing root_node_id", count),
	})

	// Orphaned nodes (no matching tree)
	db.Raw(`SELECT COUNT(*) FROM recipe_nodes rn
		LEFT JOIN recipe_trees rt ON rn.tree_id = rt.id
		WHERE rt.id IS NULL`).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Orphaned nodes (no tree)",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d orphaned nodes", count),
	})

	// Users without subscriptions
	db.Raw(`SELECT COUNT(*) FROM users u
		LEFT JOIN subscriptions s ON s.user_id = u.id
		WHERE s.id IS NULL AND u.deleted_at IS NULL`).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Users without subscriptions",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d users missing subscription", count),
	})

	// Users without settings
	db.Raw(`SELECT COUNT(*) FROM users u
		LEFT JOIN user_settings us ON us.user_id = u.id
		WHERE us.id IS NULL AND u.deleted_at IS NULL`).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Users without settings",
		Status:  warnStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d users missing settings", count),
	})

	// Users without personalization
	db.Raw(`SELECT COUNT(*) FROM users u
		LEFT JOIN personalizations p ON p.user_id = u.id
		WHERE p.id IS NULL AND u.deleted_at IS NULL`).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Users without personalization",
		Status:  warnStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d users missing personalization", count),
	})

	// Stale search cache (>48h with high hits)
	fortyEightHoursAgo := now.Add(-48 * time.Hour)
	db.Raw("SELECT COUNT(*) FROM search_caches WHERE fetched_at < ? AND hit_count > 20", fortyEightHoursAgo).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Stale high-hit search cache entries",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d stale entries with >20 hits", count),
	})

	// Embeddings missing on old recipes
	oneDayAgo := now.Add(-24 * time.Hour)
	db.Raw("SELECT COUNT(*) FROM recipes WHERE embedding IS NULL AND deleted_at IS NULL AND created_at < ?", oneDayAgo).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Missing embeddings on old recipes",
		Status:  warnStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d recipes >1d old without embeddings", count),
	})

	// Subscription counters negative
	db.Raw("SELECT COUNT(*) FROM subscriptions WHERE allergen_analyses_used < 0 OR web_searches_used < 0 OR ai_generations_used < 0").Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Negative subscription counters",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d subscriptions with negative counters", count),
	})

	// Duplicate normalized queries in search cache
	db.Raw("SELECT COUNT(*) FROM (SELECT normalized_query FROM search_caches GROUP BY normalized_query HAVING COUNT(*) > 1) sub").Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Duplicate search cache queries",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d duplicate normalized queries", count),
	})

	// Recipes with image_prompt but no image
	db.Raw("SELECT COUNT(*) FROM recipes WHERE image_prompt != '' AND (image_url = '' OR image_url IS NULL) AND deleted_at IS NULL").Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Recipes with prompt but no image",
		Status:  warnStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d recipes have image_prompt but no image_url", count),
	})

	// Orphaned canonical references
	db.Raw(`SELECT COUNT(*) FROM recipes
		WHERE canonical_id IS NOT NULL
		AND canonical_id NOT IN (SELECT id FROM canonical_recipes)
		AND deleted_at IS NULL`).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Orphaned canonical references",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d recipes reference missing canonicals", count),
	})

	// Duplicate normalized URLs in canonical cache
	db.Raw("SELECT COUNT(*) FROM (SELECT normalized_url FROM canonical_recipes GROUP BY normalized_url HAVING COUNT(*) > 1) sub").Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Duplicate canonical URLs",
		Status:  boolStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d duplicate normalized URLs", count),
	})

	// Stale canonical entries with high hits (>14d, >10 hits)
	fourteenDaysAgo := now.Add(-14 * 24 * time.Hour)
	db.Raw("SELECT COUNT(*) FROM canonical_recipes WHERE fetched_at < ? AND hit_count > 10", fourteenDaysAgo).Row().Scan(&count)
	checks = append(checks, HealthCheck{
		Name:    "Stale high-hit canonical entries",
		Status:  warnStatus(count == 0),
		Count:   count,
		Message: fmt.Sprintf("%d stale canonical entries with >10 hits", count),
	})

	return checks
}

func boolStatus(pass bool) string {
	if pass {
		return "pass"
	}
	return "fail"
}

func warnStatus(pass bool) string {
	if pass {
		return "pass"
	}
	return "warn"
}
