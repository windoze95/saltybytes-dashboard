package queries

import (
	"gorm.io/gorm"
)

type SubscriptionMetrics struct {
	TierDistribution     []LabelCount `json:"tier_distribution"`
	AvgAllergenUsed      float64      `json:"avg_allergen_used"`
	AvgSearchesUsed      float64      `json:"avg_searches_used"`
	AvgAIGenerationsUsed float64      `json:"avg_ai_generations_used"`
	UsersNearAllergenLimit int64      `json:"users_near_allergen_limit"`
	UsersNearSearchLimit   int64      `json:"users_near_search_limit"`
	UsersNearAIGenLimit    int64      `json:"users_near_ai_gen_limit"`
	UsersAtLimit         []UserAtLimit `json:"users_at_limit"`
	AllergenDistribution []LabelCount  `json:"allergen_distribution"`
	SearchDistribution   []LabelCount  `json:"search_distribution"`
	AIGenDistribution    []LabelCount  `json:"ai_gen_distribution"`
}

type UserAtLimit struct {
	UserID   uint   `json:"user_id" gorm:"column:user_id"`
	Username string `json:"username" gorm:"column:username"`
	Tier     string `json:"tier" gorm:"column:tier"`
	Allergen int    `json:"allergen_analyses_used" gorm:"column:allergen_analyses_used"`
	Search   int    `json:"web_searches_used" gorm:"column:web_searches_used"`
	AIGen    int    `json:"ai_generations_used" gorm:"column:ai_generations_used"`
}

func GetSubscriptionMetrics(db *gorm.DB) (*SubscriptionMetrics, error) {
	m := &SubscriptionMetrics{}

	// Tier distribution
	db.Table("subscriptions").
		Select("tier as label, COUNT(*) as count").
		Group("tier").
		Find(&m.TierDistribution)

	// Average usage for free tier
	safeScan(db.Table("subscriptions").Where("tier = 'free'").Select("COALESCE(AVG(allergen_analyses_used), 0)"), &m.AvgAllergenUsed)
	safeScan(db.Table("subscriptions").Where("tier = 'free'").Select("COALESCE(AVG(web_searches_used), 0)"), &m.AvgSearchesUsed)
	safeScan(db.Table("subscriptions").Where("tier = 'free'").Select("COALESCE(AVG(ai_generations_used), 0)"), &m.AvgAIGenerationsUsed)

	// Users near limits (>= 18 out of 20)
	db.Table("subscriptions").
		Where("tier = 'free' AND allergen_analyses_used >= 18").
		Count(&m.UsersNearAllergenLimit)

	db.Table("subscriptions").
		Where("tier = 'free' AND web_searches_used >= 18").
		Count(&m.UsersNearSearchLimit)

	db.Table("subscriptions").
		Where("tier = 'free' AND ai_generations_used >= 18").
		Count(&m.UsersNearAIGenLimit)

	// Users at limits (sortable table)
	db.Raw(`SELECT s.user_id, u.username, s.tier,
		s.allergen_analyses_used, s.web_searches_used, s.ai_generations_used
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		WHERE s.tier = 'free' AND (
			s.allergen_analyses_used >= 15 OR
			s.web_searches_used >= 15 OR
			s.ai_generations_used >= 15
		)
		ORDER BY (s.allergen_analyses_used + s.web_searches_used + s.ai_generations_used) DESC
		LIMIT 50`).Find(&m.UsersAtLimit)

	// Usage distributions (buckets)
	db.Raw(`SELECT
		CASE
			WHEN allergen_analyses_used = 0 THEN '0'
			WHEN allergen_analyses_used BETWEEN 1 AND 5 THEN '1-5'
			WHEN allergen_analyses_used BETWEEN 6 AND 10 THEN '6-10'
			WHEN allergen_analyses_used BETWEEN 11 AND 15 THEN '11-15'
			WHEN allergen_analyses_used BETWEEN 16 AND 20 THEN '16-20'
			ELSE '20+'
		END as label,
		COUNT(*) as count
		FROM subscriptions WHERE tier = 'free'
		GROUP BY label
		ORDER BY MIN(allergen_analyses_used)`).Find(&m.AllergenDistribution)

	db.Raw(`SELECT
		CASE
			WHEN web_searches_used = 0 THEN '0'
			WHEN web_searches_used BETWEEN 1 AND 5 THEN '1-5'
			WHEN web_searches_used BETWEEN 6 AND 10 THEN '6-10'
			WHEN web_searches_used BETWEEN 11 AND 15 THEN '11-15'
			WHEN web_searches_used BETWEEN 16 AND 20 THEN '16-20'
			ELSE '20+'
		END as label,
		COUNT(*) as count
		FROM subscriptions WHERE tier = 'free'
		GROUP BY label
		ORDER BY MIN(web_searches_used)`).Find(&m.SearchDistribution)

	db.Raw(`SELECT
		CASE
			WHEN ai_generations_used = 0 THEN '0'
			WHEN ai_generations_used BETWEEN 1 AND 5 THEN '1-5'
			WHEN ai_generations_used BETWEEN 6 AND 10 THEN '6-10'
			WHEN ai_generations_used BETWEEN 11 AND 15 THEN '11-15'
			WHEN ai_generations_used BETWEEN 16 AND 20 THEN '16-20'
			ELSE '20+'
		END as label,
		COUNT(*) as count
		FROM subscriptions WHERE tier = 'free'
		GROUP BY label
		ORDER BY MIN(ai_generations_used)`).Find(&m.AIGenDistribution)

	return m, nil
}
