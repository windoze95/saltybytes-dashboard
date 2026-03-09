package queries

import (
	"time"

	"gorm.io/gorm"
)

type RecipeMetrics struct {
	TotalRecipes           int64            `json:"total_recipes"`
	RecipesToday           int64            `json:"recipes_today"`
	RecipesThisWeek        int64            `json:"recipes_this_week"`
	RecipesWithImages      int64            `json:"recipes_with_images"`
	RecipesWithEmbeddings  int64            `json:"recipes_with_embeddings"`
	RecipesWithSourceURL   int64            `json:"recipes_with_source_url"`
	EmbeddingCoverage      float64          `json:"embedding_coverage"`
	ImageCoverage          float64          `json:"image_coverage"`
	UserEditedRate         float64          `json:"user_edited_rate"`
	UserEditedCount        int64            `json:"user_edited_count"`
	AvgIngredientsPerRecipe float64         `json:"avg_ingredients_per_recipe"`
	AvgCookTime            float64          `json:"avg_cook_time"`
	AvgPortions            float64          `json:"avg_portions"`
	ImageRegenCount        int64            `json:"image_regen_count"`
	DeletedToday           int64            `json:"deleted_today"`
	DailyCreation          []DailyLabelCount `json:"daily_creation"`
	NodeTypeDistribution   []LabelCount     `json:"node_type_distribution"`
	TopHashtags            []TopEntry       `json:"top_hashtags"`
	RecipesPerUser         []LabelCount     `json:"recipes_per_user"`
	ImportBreakdown        []LabelCount     `json:"import_breakdown"`

	// Canonical
	CanonicalLinked       int64    `json:"canonical_linked"`
	CanonicalNonDiverged  int64    `json:"canonical_non_diverged"`
	CanonicalDiverged     int64    `json:"canonical_diverged"`
	DivergenceRate        float64  `json:"divergence_rate"`

	// Fork chains
	MaxForkDepth          int64    `json:"max_fork_depth"`
	AvgForkDepth          float64  `json:"avg_fork_depth"`
	ForkCount             int64    `json:"fork_count"`

	// Trees
	TotalTrees            int64    `json:"total_trees"`
	EphemeralNodes        int64    `json:"ephemeral_nodes"`
}

func GetRecipeMetrics(db *gorm.DB) (*RecipeMetrics, error) {
	m := &RecipeMetrics{}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := today.AddDate(0, 0, -7)

	db.Table("recipes").Where("deleted_at IS NULL").Count(&m.TotalRecipes)
	db.Table("recipes").Where("deleted_at IS NULL AND created_at >= ?", today).Count(&m.RecipesToday)
	db.Table("recipes").Where("deleted_at IS NULL AND created_at >= ?", weekAgo).Count(&m.RecipesThisWeek)
	db.Table("recipes").Where("deleted_at IS NULL AND image_url != ''").Count(&m.RecipesWithImages)
	db.Table("recipes").Where("deleted_at IS NULL AND embedding IS NOT NULL").Count(&m.RecipesWithEmbeddings)
	db.Table("recipes").Where("deleted_at IS NULL AND user_edited = true").Count(&m.UserEditedCount)
	db.Table("recipes").Where("deleted_at IS NOT NULL AND deleted_at >= ?", today).Count(&m.DeletedToday)

	if m.TotalRecipes > 0 {
		m.EmbeddingCoverage = float64(m.RecipesWithEmbeddings) / float64(m.TotalRecipes) * 100
		m.ImageCoverage = float64(m.RecipesWithImages) / float64(m.TotalRecipes) * 100
		m.UserEditedRate = float64(m.UserEditedCount) / float64(m.TotalRecipes) * 100
	}

	// These columns may not exist on every schema version — use safeScan
	safeScan(db.Raw("SELECT COUNT(*) FROM recipes WHERE deleted_at IS NULL AND source_url IS NOT NULL AND source_url != ''"), &m.RecipesWithSourceURL)
	safeScan(db.Raw("SELECT COUNT(*) FROM recipes WHERE deleted_at IS NULL AND original_image_url IS NOT NULL AND original_image_url != ''"), &m.ImageRegenCount)
	safeScan(db.Raw("SELECT COALESCE(AVG(cook_time), 0) FROM recipes WHERE deleted_at IS NULL AND cook_time > 0"), &m.AvgCookTime)
	safeScan(db.Raw("SELECT COALESCE(AVG(portions), 0) FROM recipes WHERE deleted_at IS NULL AND portions > 0"), &m.AvgPortions)
	safeScan(db.Raw("SELECT COALESCE(AVG(jsonb_array_length(ingredients)), 0) FROM recipes WHERE deleted_at IS NULL AND ingredients IS NOT NULL AND ingredients != 'null'"), &m.AvgIngredientsPerRecipe)

	// Node type distribution
	db.Table("recipe_nodes").
		Select("type as label, COUNT(*) as count").
		Group("type").
		Order("count DESC").
		Find(&m.NodeTypeDistribution)

	// Daily creation by type (last 30 days)
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	db.Table("recipe_nodes").
		Select("DATE(created_at) as date, type as label, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at), type").
		Order("date").
		Find(&m.DailyCreation)

	// Import breakdown
	db.Table("recipe_nodes").
		Select("type as label, COUNT(*) as count").
		Where("type IN ('import_link','import_vision','import_text','user_input')").
		Group("type").
		Find(&m.ImportBreakdown)

	// Top hashtags
	db.Raw(`SELECT t.hashtag as label, COUNT(*) as count
		FROM tags t
		JOIN recipe_tags rt ON rt.tag_id = t.id
		GROUP BY t.hashtag
		ORDER BY count DESC
		LIMIT 20`).Find(&m.TopHashtags)

	// Recipes per user distribution (bucket into ranges)
	db.Raw(`SELECT
		CASE
			WHEN cnt = 1 THEN '1'
			WHEN cnt BETWEEN 2 AND 5 THEN '2-5'
			WHEN cnt BETWEEN 6 AND 10 THEN '6-10'
			WHEN cnt BETWEEN 11 AND 25 THEN '11-25'
			WHEN cnt BETWEEN 26 AND 50 THEN '26-50'
			ELSE '50+'
		END as label,
		COUNT(*) as count
		FROM (SELECT created_by_id, COUNT(*) as cnt FROM recipes WHERE deleted_at IS NULL GROUP BY created_by_id) sub
		GROUP BY label
		ORDER BY MIN(cnt)`).Find(&m.RecipesPerUser)

	// Canonical stats
	db.Table("recipes").Where("deleted_at IS NULL AND canonical_id IS NOT NULL").Count(&m.CanonicalLinked)
	db.Table("recipes").Where("deleted_at IS NULL AND canonical_id IS NOT NULL AND has_diverged = false").Count(&m.CanonicalNonDiverged)
	db.Table("recipes").Where("deleted_at IS NULL AND canonical_id IS NOT NULL AND has_diverged = true").Count(&m.CanonicalDiverged)
	if m.CanonicalLinked > 0 {
		m.DivergenceRate = float64(m.CanonicalDiverged) / float64(m.CanonicalLinked) * 100
	}

	// Fork stats
	db.Table("recipes").Where("deleted_at IS NULL AND forked_from_id IS NOT NULL").Count(&m.ForkCount)

	// Fork chain depth via recursive CTE
	safeScan(db.Raw(`WITH RECURSIVE fork_chain AS (
		SELECT id, forked_from_id, 1 as depth FROM recipes WHERE forked_from_id IS NOT NULL AND deleted_at IS NULL
		UNION ALL
		SELECT r.id, r.forked_from_id, fc.depth + 1
		FROM recipes r
		JOIN fork_chain fc ON r.forked_from_id = fc.id
		WHERE r.deleted_at IS NULL
	)
	SELECT COALESCE(MAX(depth), 0), COALESCE(AVG(depth), 0) FROM fork_chain`), &m.MaxForkDepth, &m.AvgForkDepth)

	// Trees
	db.Table("recipe_trees").Count(&m.TotalTrees)
	db.Table("recipe_nodes").Where("is_ephemeral = true").Count(&m.EphemeralNodes)

	return m, nil
}
