package cache

import (
	"log"
	"sync"
	"time"

	"github.com/saltybytes/saltybytes-dashboard/internal/queries"
	"github.com/saltybytes/saltybytes-dashboard/internal/ratecard"
	"gorm.io/gorm"
)

type MetricCache struct {
	mu sync.RWMutex
	db *gorm.DB
	rc *ratecard.RateCard

	Users          *queries.UserMetrics
	Recipes        *queries.RecipeMetrics
	Canonical      *queries.CanonicalMetrics
	SearchCache    *queries.SearchCacheMetrics
	Subscriptions  *queries.SubscriptionMetrics
	Allergens      *queries.AllergenMetrics
	Families       *queries.FamilyMetrics
	Infrastructure *queries.InfrastructureMetrics
	HealthChecks   []queries.HealthCheck
	Overview       *OverviewMetrics
	CostCenter     *CostCenterMetrics

	lastRefresh time.Time
}

type OverviewMetrics struct {
	TotalUsers            int64   `json:"total_users"`
	TotalRecipes          int64   `json:"total_recipes"`
	TotalCanonicals       int64   `json:"total_canonicals"`
	TodayEstimatedCost    float64 `json:"today_estimated_cost"`
	MonthToDateCost       float64 `json:"month_to_date_cost"`
	MonthProjection       float64 `json:"month_projection"`
	SearchCacheHitRate    float64 `json:"search_cache_hit_rate"`
	FirecrawlCreditsUsed  int64   `json:"firecrawl_credits_used"`
	FirecrawlCreditsLeft  int     `json:"firecrawl_credits_left"`
	HealthChecksPassing   int     `json:"health_checks_passing"`
	HealthChecksTotal     int     `json:"health_checks_total"`
	RecipesCreatedWeek    []queries.DailyCount `json:"recipes_created_week"`
}

type CostCenterMetrics struct {
	DailyCostByProvider  []queries.DailyLabelCount `json:"daily_cost_by_provider"`
	CostByFeature        []queries.LabelCount      `json:"cost_by_feature"`
	CostPerUser          float64                   `json:"cost_per_user"`
	CostPerRecipe        float64                   `json:"cost_per_recipe"`
	SearchCacheSavings   float64                   `json:"search_cache_savings"`
	CanonicalCacheSavings float64                  `json:"canonical_cache_savings"`
	AllergenCacheSavings float64                   `json:"allergen_cache_savings"`
	TotalSavings         float64                   `json:"total_savings"`
	FirecrawlCreditsUsed int64                     `json:"firecrawl_credits_used"`
	FirecrawlCreditsMax  int                       `json:"firecrawl_credits_max"`
	MonthlyFixedCosts    float64                   `json:"monthly_fixed_costs"`
	RateCard             ratecard.RateCard         `json:"rate_card"`
}

func New(db *gorm.DB, rc *ratecard.RateCard) *MetricCache {
	return &MetricCache{db: db, rc: rc}
}

func (c *MetricCache) Start() {
	c.Refresh()

	// Light metrics every 2 minutes
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			c.Refresh()
		}
	}()
}

func (c *MetricCache) Refresh() {
	log.Println("Refreshing metric cache...")
	start := time.Now()

	var wg sync.WaitGroup
	var (
		users          *queries.UserMetrics
		recipes        *queries.RecipeMetrics
		canonical      *queries.CanonicalMetrics
		searchCache    *queries.SearchCacheMetrics
		subscriptions  *queries.SubscriptionMetrics
		allergens      *queries.AllergenMetrics
		families       *queries.FamilyMetrics
		infrastructure *queries.InfrastructureMetrics
		healthChecks   []queries.HealthCheck
	)

	rc := c.rc.Get()

	wg.Add(8)

	go func() { defer wg.Done(); users, _ = queries.GetUserMetrics(c.db) }()
	go func() { defer wg.Done(); recipes, _ = queries.GetRecipeMetrics(c.db) }()
	go func() { defer wg.Done(); canonical, _ = queries.GetCanonicalMetrics(c.db) }()
	go func() { defer wg.Done(); searchCache, _ = queries.GetSearchCacheMetrics(c.db) }()
	go func() { defer wg.Done(); subscriptions, _ = queries.GetSubscriptionMetrics(c.db) }()
	go func() { defer wg.Done(); allergens, _ = queries.GetAllergenMetrics(c.db) }()
	go func() { defer wg.Done(); families, _ = queries.GetFamilyMetrics(c.db) }()
	go func() {
		defer wg.Done()
		infrastructure, _ = queries.GetInfrastructureMetrics(c.db, rc.AWSS3PerGB)
	}()

	wg.Wait()

	healthChecks = queries.GetHealthChecks(c.db)

	// Compute overview
	overview := &OverviewMetrics{}
	if users != nil {
		overview.TotalUsers = users.TotalUsers
	}
	if recipes != nil {
		overview.TotalRecipes = recipes.TotalRecipes
	}
	if canonical != nil {
		overview.TotalCanonicals = canonical.TotalEntries
		overview.FirecrawlCreditsUsed = canonical.FirecrawlCallsThisMonth
		overview.FirecrawlCreditsLeft = rc.FirecrawlMonthlyCredits - int(canonical.FirecrawlCallsThisMonth)
		if overview.FirecrawlCreditsLeft < 0 {
			overview.FirecrawlCreditsLeft = 0
		}
	}

	// Health checks summary
	overview.HealthChecksTotal = len(healthChecks)
	for _, hc := range healthChecks {
		if hc.Status == "pass" {
			overview.HealthChecksPassing++
		}
	}

	// Compute cost center
	costCenter := &CostCenterMetrics{
		RateCard:         rc,
		FirecrawlCreditsMax: rc.FirecrawlMonthlyCredits,
	}
	if canonical != nil {
		costCenter.FirecrawlCreditsUsed = canonical.FirecrawlCallsThisMonth
		// Canonical cache savings: each avoided Haiku call saves ~$0.01 estimated
		haikuCostPerExtraction := (rc.AnthropicHaikuInputPerMTok*2000 + rc.AnthropicHaikuOutputPerMTok*3000) / 1_000_000
		costCenter.CanonicalCacheSavings = float64(canonical.TotalAIExtractionsSaved) * haikuCostPerExtraction
	}
	costCenter.MonthlyFixedCosts = rc.AWSRDSMonthly + rc.AWSECSMonthly + rc.BraveMonthlyPlan
	if infrastructure != nil {
		costCenter.MonthlyFixedCosts += infrastructure.S3EstimatedCost
	}
	costCenter.TotalSavings = costCenter.SearchCacheSavings + costCenter.CanonicalCacheSavings + costCenter.AllergenCacheSavings

	if users != nil && users.TotalUsers > 0 {
		costCenter.CostPerUser = costCenter.MonthlyFixedCosts / float64(users.TotalUsers)
	}
	if recipes != nil && recipes.TotalRecipes > 0 {
		costCenter.CostPerRecipe = costCenter.MonthlyFixedCosts / float64(recipes.TotalRecipes)
	}

	// Store everything
	c.mu.Lock()
	c.Users = users
	c.Recipes = recipes
	c.Canonical = canonical
	c.SearchCache = searchCache
	c.Subscriptions = subscriptions
	c.Allergens = allergens
	c.Families = families
	c.Infrastructure = infrastructure
	c.HealthChecks = healthChecks
	c.Overview = overview
	c.CostCenter = costCenter
	c.lastRefresh = time.Now()
	c.mu.Unlock()

	log.Printf("Metric cache refreshed in %v", time.Since(start))
}

func (c *MetricCache) GetUsers() *queries.UserMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Users
}

func (c *MetricCache) GetRecipes() *queries.RecipeMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Recipes
}

func (c *MetricCache) GetCanonical() *queries.CanonicalMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Canonical
}

func (c *MetricCache) GetSearchCache() *queries.SearchCacheMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SearchCache
}

func (c *MetricCache) GetSubscriptions() *queries.SubscriptionMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Subscriptions
}

func (c *MetricCache) GetAllergens() *queries.AllergenMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Allergens
}

func (c *MetricCache) GetFamilies() *queries.FamilyMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Families
}

func (c *MetricCache) GetInfrastructure() *queries.InfrastructureMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Infrastructure
}

func (c *MetricCache) GetHealthChecks() []queries.HealthCheck {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HealthChecks
}

func (c *MetricCache) GetOverview() *OverviewMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Overview
}

func (c *MetricCache) GetCostCenter() *CostCenterMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CostCenter
}

func (c *MetricCache) GetLastRefresh() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastRefresh
}
