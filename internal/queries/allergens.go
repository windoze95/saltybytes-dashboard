package queries

import (
	"time"

	"gorm.io/gorm"
)

type AllergenMetrics struct {
	TotalAnalyses      int64        `json:"total_analyses"`
	AnalysesToday      int64        `json:"analyses_today"`
	AvgConfidence      float64      `json:"avg_confidence"`
	RequiresReviewRate float64      `json:"requires_review_rate"`
	RequiresReviewCount int64       `json:"requires_review_count"`
	PremiumCount       int64        `json:"premium_count"`
	FreeCount          int64        `json:"free_count"`
	AllergenFlags      []LabelCount `json:"allergen_flags"`
	DailyVolume        []DailyCount `json:"daily_volume"`
	ConfidenceDist     []LabelCount `json:"confidence_distribution"`
}

func GetAllergenMetrics(db *gorm.DB) (*AllergenMetrics, error) {
	m := &AllergenMetrics{}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	db.Table("allergen_analyses").Count(&m.TotalAnalyses)
	db.Table("allergen_analyses").Where("created_at >= ?", today).Count(&m.AnalysesToday)
	db.Table("allergen_analyses").Where("requires_review = true").Count(&m.RequiresReviewCount)
	db.Table("allergen_analyses").Where("is_premium = true").Count(&m.PremiumCount)
	db.Table("allergen_analyses").Where("is_premium = false").Count(&m.FreeCount)

	if m.TotalAnalyses > 0 {
		m.RequiresReviewRate = float64(m.RequiresReviewCount) / float64(m.TotalAnalyses) * 100
	}

	db.Table("allergen_analyses").
		Select("COALESCE(AVG(confidence), 0)").
		Row().Scan(&m.AvgConfidence)

	// Allergen flags breakdown
	db.Raw(`SELECT 'Nuts' as label, COUNT(*) as count FROM allergen_analyses WHERE contains_nuts = true
		UNION ALL SELECT 'Dairy', COUNT(*) FROM allergen_analyses WHERE contains_dairy = true
		UNION ALL SELECT 'Gluten', COUNT(*) FROM allergen_analyses WHERE contains_gluten = true
		UNION ALL SELECT 'Soy', COUNT(*) FROM allergen_analyses WHERE contains_soy = true
		UNION ALL SELECT 'Seed Oils', COUNT(*) FROM allergen_analyses WHERE contains_seed_oils = true
		UNION ALL SELECT 'Shellfish', COUNT(*) FROM allergen_analyses WHERE contains_shellfish = true
		UNION ALL SELECT 'Eggs', COUNT(*) FROM allergen_analyses WHERE contains_eggs = true
		ORDER BY count DESC`).Find(&m.AllergenFlags)

	// Daily volume (30 days)
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	db.Table("allergen_analyses").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailyVolume)

	// Confidence distribution
	db.Raw(`SELECT
		CASE
			WHEN confidence >= 0.9 THEN 'High (90-100%)'
			WHEN confidence >= 0.7 THEN 'Medium (70-89%)'
			WHEN confidence >= 0.5 THEN 'Low (50-69%)'
			ELSE 'Very Low (<50%)'
		END as label,
		COUNT(*) as count
		FROM allergen_analyses
		GROUP BY label
		ORDER BY MIN(confidence) DESC`).Find(&m.ConfidenceDist)

	return m, nil
}
