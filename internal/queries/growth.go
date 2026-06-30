package queries

import (
	"time"

	"gorm.io/gorm"
)

// GrowthMetrics captures growth trends, the signup -> first-recipe activation
// funnel, and engagement/retention proxies drawn across the users, recipes and
// subscriptions tables. It deliberately leans on time-series and cross-table
// engagement rather than the per-entity counts already on the Users/Recipes
// pages.
type GrowthMetrics struct {
	// Headline
	TotalUsers           int64   `json:"total_users"`
	NewUsersToday        int64   `json:"new_users_today"`
	NewUsers7d           int64   `json:"new_users_7d"`
	NewUsers30d          int64   `json:"new_users_30d"`
	TotalRecipes         int64   `json:"total_recipes"`
	RecipesPerActiveUser float64 `json:"recipes_per_active_user"`

	// Activation funnel (signup -> first recipe)
	UsersWithRecipes int64   `json:"users_with_recipes"`
	ActivationRate   float64 `json:"activation_rate"`

	// Time series (last 30 days)
	DailySignups    []DailyCount `json:"daily_signups"`
	CumulativeUsers []DailyCount `json:"cumulative_users"`
	DailyRecipes    []DailyCount `json:"daily_recipes"`

	// Engagement + monetization mix
	RecipesPerUser   []LabelCount `json:"recipes_per_user"`
	TierDistribution []LabelCount `json:"tier_distribution"`
}

// GetGrowthMetrics computes growth/engagement metrics. It is drift-tolerant:
// .Count()/.Find() silently set tx.Error (leaving the destination at its zero
// value) and scalar reads go through safeScan, so a missing table or column
// leaves the affected metric zero/empty rather than panicking. Soft-deletes are
// respected (deleted_at IS NULL) consistent with users.go / recipes.go.
func GetGrowthMetrics(db *gorm.DB) (*GrowthMetrics, error) {
	m := &GrowthMetrics{}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := today.AddDate(0, 0, -7)
	thirtyDaysAgo := today.AddDate(0, 0, -30)

	// Headline counts
	db.Table("users").Where("deleted_at IS NULL").Count(&m.TotalUsers)
	db.Table("users").Where("deleted_at IS NULL AND created_at >= ?", today).Count(&m.NewUsersToday)
	db.Table("users").Where("deleted_at IS NULL AND created_at >= ?", weekAgo).Count(&m.NewUsers7d)
	db.Table("users").Where("deleted_at IS NULL AND created_at >= ?", thirtyDaysAgo).Count(&m.NewUsers30d)
	db.Table("recipes").Where("deleted_at IS NULL").Count(&m.TotalRecipes)

	if m.TotalUsers > 0 {
		m.RecipesPerActiveUser = float64(m.TotalRecipes) / float64(m.TotalUsers)
	}

	// Activation funnel: distinct active users with >= 1 non-deleted recipe.
	safeScan(db.Raw(`SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN recipes r ON r.created_by_id = u.id AND r.deleted_at IS NULL
		WHERE u.deleted_at IS NULL`), &m.UsersWithRecipes)
	if m.TotalUsers > 0 {
		m.ActivationRate = float64(m.UsersWithRecipes) / float64(m.TotalUsers) * 100
	}

	// Daily signups (last 30 days)
	db.Table("users").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("deleted_at IS NULL AND created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailySignups)

	// Cumulative users over time: running total within the 30-day window,
	// offset by everyone who signed up before it so the line starts at the
	// real total instead of zero. Cast to bigint so the numeric SUM scans into
	// int64 cleanly.
	var baseline int64
	safeScan(db.Raw("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND created_at < ?", thirtyDaysAgo), &baseline)
	db.Raw(`SELECT date, (? + SUM(daily) OVER (ORDER BY date))::bigint AS count
		FROM (
			SELECT DATE(created_at) AS date, COUNT(*) AS daily
			FROM users
			WHERE deleted_at IS NULL AND created_at >= ?
			GROUP BY DATE(created_at)
		) d
		ORDER BY date`, baseline, thirtyDaysAgo).Find(&m.CumulativeUsers)

	// Daily recipes created (last 30 days)
	db.Table("recipes").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("deleted_at IS NULL AND created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailyRecipes)

	// Engagement distribution: recipes-per-user buckets. The LEFT JOIN keeps
	// users with zero recipes so the '0' bucket is represented.
	db.Raw(`SELECT
		CASE
			WHEN cnt = 0 THEN '0'
			WHEN cnt = 1 THEN '1'
			WHEN cnt BETWEEN 2 AND 5 THEN '2-5'
			WHEN cnt BETWEEN 6 AND 20 THEN '6-20'
			ELSE '20+'
		END AS label,
		COUNT(*) AS count
		FROM (
			SELECT u.id, COUNT(r.id) AS cnt
			FROM users u
			LEFT JOIN recipes r ON r.created_by_id = u.id AND r.deleted_at IS NULL
			WHERE u.deleted_at IS NULL
			GROUP BY u.id
		) sub
		GROUP BY label
		ORDER BY MIN(cnt)`).Find(&m.RecipesPerUser)

	// Subscription tier mix (free vs premium counts)
	db.Table("subscriptions").
		Select("tier as label, COUNT(*) as count").
		Group("tier").
		Order("count DESC").
		Find(&m.TierDistribution)

	return m, nil
}
