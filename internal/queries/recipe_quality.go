package queries

import (
	"time"

	"gorm.io/gorm"
)

// DailyRecipeOutcome is the per-day split of recipe creations into successful
// (status = 'ready') vs failed (status = 'failed') outcomes.
type DailyRecipeOutcome struct {
	Date      time.Time `json:"date" gorm:"column:date"`
	Succeeded int64     `json:"succeeded" gorm:"column:succeeded"`
	Failed    int64     `json:"failed" gorm:"column:failed"`
}

// RecipeQualityMetrics describes extraction quality (free vs paid AI) and recipe
// completeness/health signals. It is drift-tolerant: any column whose existence
// is uncertain is guarded with hasColumn so a missing column leaves the relevant
// metric zero/empty instead of crashing.
type RecipeQualityMetrics struct {
	// Headline
	TotalRecipes        int64   `json:"total_recipes"`
	FailedRecipes       int64   `json:"failed_recipes"`
	FailureRate         float64 `json:"failure_rate"`
	TotalCanonical      int64   `json:"total_canonical"`
	FreeExtractionCount int64   `json:"free_extraction_count"`
	PaidExtractionCount int64   `json:"paid_extraction_count"`
	FreeExtractionRate  float64 `json:"free_extraction_rate"`
	PaidExtractionRate  float64 `json:"paid_extraction_rate"`

	// Distributions
	ExtractionMethodDist []LabelCount `json:"extraction_method_distribution"`
	StatusDistribution   []LabelCount `json:"status_distribution"`
	TypeDistribution     []LabelCount `json:"type_distribution"` // empty if recipes has no type column

	// Completeness (guarded; stays zero if the jsonb/array columns are absent)
	AvgIngredients            float64 `json:"avg_ingredients"`
	AvgSteps                  float64 `json:"avg_steps"`
	RecipesMissingIngredients int64   `json:"recipes_missing_ingredients"`

	// Creation outcome over time (last 30 days)
	DailyOutcome []DailyRecipeOutcome `json:"daily_outcome"`
}

// GetRecipeQuality aggregates extraction-quality and recipe-completeness signals.
// Soft-deleted recipes are excluded (deleted_at IS NULL) consistent with recipes.go;
// canonical_recipes are counted in full consistent with canonical.go.
func GetRecipeQuality(db *gorm.DB) (*RecipeQualityMetrics, error) {
	m := &RecipeQualityMetrics{}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Headline counts
	db.Table("recipes").Where("deleted_at IS NULL").Count(&m.TotalRecipes)
	db.Table("recipes").Where("deleted_at IS NULL AND status = 'failed'").Count(&m.FailedRecipes)
	if m.TotalRecipes > 0 {
		m.FailureRate = float64(m.FailedRecipes) / float64(m.TotalRecipes) * 100
	}

	// Free vs paid extraction (the free-vs-paid story) from canonical_recipes.
	db.Table("canonical_recipes").Count(&m.TotalCanonical)
	db.Table("canonical_recipes").Where("extraction_method IN ('json_ld','firecrawl_json_ld')").Count(&m.FreeExtractionCount)
	db.Table("canonical_recipes").Where("extraction_method IN ('haiku','firecrawl_haiku')").Count(&m.PaidExtractionCount)
	if m.TotalCanonical > 0 {
		m.FreeExtractionRate = float64(m.FreeExtractionCount) / float64(m.TotalCanonical) * 100
		m.PaidExtractionRate = float64(m.PaidExtractionCount) / float64(m.TotalCanonical) * 100
	}

	// Extraction method distribution
	db.Table("canonical_recipes").
		Select("extraction_method as label, COUNT(*) as count").
		Group("extraction_method").
		Order("count DESC").
		Find(&m.ExtractionMethodDist)

	// Recipe status distribution (ready / generating / failed / ...)
	db.Table("recipes").
		Where("deleted_at IS NULL").
		Select("status as label, COUNT(*) as count").
		Group("status").
		Order("count DESC").
		Find(&m.StatusDistribution)

	// Recipe source/type mix — only if a type column exists on recipes. The
	// recipes table may NOT have one (the type lives on recipe_nodes), so guard
	// and omit gracefully when absent.
	typeCol := ""
	if hasColumn(db, "recipes", "recipe_type") {
		typeCol = "recipe_type"
	} else if hasColumn(db, "recipes", "type") {
		typeCol = "type"
	}
	if typeCol != "" {
		db.Table("recipes").
			Where("deleted_at IS NULL").
			Select(typeCol + " as label, COUNT(*) as count").
			Group(typeCol).
			Order("count DESC").
			Find(&m.TypeDistribution)
	}

	// Completeness from the recipe definition. ingredients is a jsonb array and
	// instructions is a text[] (confirmed from the GORM model) — both guarded.
	if hasColumn(db, "recipes", "ingredients") {
		safeScan(db.Raw("SELECT COALESCE(AVG(jsonb_array_length(ingredients)), 0) FROM recipes WHERE deleted_at IS NULL AND ingredients IS NOT NULL AND ingredients != 'null'"), &m.AvgIngredients)
		safeScan(db.Raw("SELECT COUNT(*) FROM recipes WHERE deleted_at IS NULL AND COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(ingredients) = 'array' THEN ingredients ELSE '[]'::jsonb END), 0) = 0"), &m.RecipesMissingIngredients)
	}
	if hasColumn(db, "recipes", "instructions") {
		safeScan(db.Raw("SELECT COALESCE(AVG(cardinality(instructions)), 0) FROM recipes WHERE deleted_at IS NULL AND instructions IS NOT NULL"), &m.AvgSteps)
	}

	// Creation outcome over time (last 30 days): succeeded vs failed per day.
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	db.Table("recipes").
		Select("DATE(created_at) as date, COUNT(*) FILTER (WHERE status = 'ready') as succeeded, COUNT(*) FILTER (WHERE status = 'failed') as failed").
		Where("deleted_at IS NULL AND created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailyOutcome)

	return m, nil
}
