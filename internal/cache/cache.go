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
	AIModels       *AIModelMetrics
	AIRegistry     *queries.AIModelRegistry
	AIOps          *queries.AIOpsMetrics
	CacheEconomics *queries.CacheEconomics
	VideoEconomics *queries.VideoEconomics
	Growth         *queries.GrowthMetrics
	RecipeQuality  *queries.RecipeQualityMetrics

	lastRefresh time.Time
}

// AIModelMetrics is the AI usage report plus counterfactual ("what model X
// would have cost") pricing of the recorded token volume.
type AIModelMetrics struct {
	*queries.AIModelUsage
	Counterfactuals []AICounterfactual `json:"counterfactuals"`
}

// AICounterfactual is what the period's recorded tokens would have cost on a
// given model. VsActualPct is negative when the candidate is cheaper than what
// was actually spent.
type AICounterfactual struct {
	Label       string  `json:"label"`
	CostUSD     float64 `json:"cost_usd"`
	VsActualPct float64 `json:"vs_actual_pct"`
}

type OverviewMetrics struct {
	TotalUsers           int64                `json:"total_users"`
	TotalRecipes         int64                `json:"total_recipes"`
	TotalCanonicals      int64                `json:"total_canonicals"`
	TodayEstimatedCost   float64              `json:"today_estimated_cost"`
	MonthToDateCost      float64              `json:"month_to_date_cost"`
	MonthProjection      float64              `json:"month_projection"`
	SearchCacheHitRate   float64              `json:"search_cache_hit_rate"`
	FirecrawlCreditsUsed int64                `json:"firecrawl_credits_used"`
	FirecrawlCreditsLeft int                  `json:"firecrawl_credits_left"`
	HealthChecksPassing  int                  `json:"health_checks_passing"`
	HealthChecksTotal    int                  `json:"health_checks_total"`
	RecipesCreatedWeek   []queries.DailyCount `json:"recipes_created_week"`
}

type CostCenterMetrics struct {
	DailyCostByProvider   []queries.DailyLabelCount `json:"daily_cost_by_provider"`
	CostByFeature         []queries.LabelCount      `json:"cost_by_feature"`
	CostPerUser           float64                   `json:"cost_per_user"`
	CostPerRecipe         float64                   `json:"cost_per_recipe"`
	SearchCacheSavings    float64                   `json:"search_cache_savings"`
	CanonicalCacheSavings float64                   `json:"canonical_cache_savings"`
	AllergenCacheSavings  float64                   `json:"allergen_cache_savings"`
	TotalSavings          float64                   `json:"total_savings"`
	FirecrawlCreditsUsed  int64                     `json:"firecrawl_credits_used"`
	FirecrawlCreditsMax   int                       `json:"firecrawl_credits_max"`
	MonthlyFixedCosts     float64                   `json:"monthly_fixed_costs"`
	RateCard              ratecard.Rates            `json:"rate_card"`
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
		aiOps          *queries.AIOpsMetrics
		cacheEcon      *queries.CacheEconomics
		videoEcon      *queries.VideoEconomics
		growth         *queries.GrowthMetrics
		recipeQuality  *queries.RecipeQualityMetrics
	)

	rc := c.rc.Get()

	wg.Add(13)

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
	go func() { defer wg.Done(); aiOps, _ = queries.GetAIOps(c.db, 30) }()
	go func() { defer wg.Done(); cacheEcon, _ = queries.GetCacheEconomics(c.db) }()
	go func() { defer wg.Done(); videoEcon, _ = queries.GetVideoEconomics(c.db, 30) }()
	go func() { defer wg.Done(); growth, _ = queries.GetGrowthMetrics(c.db) }()
	go func() { defer wg.Done(); recipeQuality, _ = queries.GetRecipeQuality(c.db) }()

	wg.Wait()

	healthChecks = queries.GetHealthChecks(c.db)

	// Light-tier model registry + active selection (read-only mirror of the
	// API's ai_model_options / ai_config). Tiny tables — read every refresh.
	aiRegistry, _ := queries.GetAIModelRegistry(c.db)

	// AI model usage (last 30 days) + counterfactual pricing of the recorded
	// token volume against each candidate model's rate.
	aiUsage, _ := queries.GetAIModelUsage(c.db, 30)
	aiModels := &AIModelMetrics{AIModelUsage: aiUsage}
	if aiUsage != nil {
		inM := float64(aiUsage.TotalInputTokens) / 1_000_000
		outM := float64(aiUsage.TotalOutputTokens) / 1_000_000
		candidates := []struct {
			label   string
			in, out float64
		}{
			{"Claude Haiku", rc.AnthropicHaikuInputPerMTok, rc.AnthropicHaikuOutputPerMTok},
			{"Claude Sonnet", rc.AnthropicSonnetInputPerMTok, rc.AnthropicSonnetOutputPerMTok},
			{"GPT-4o mini", rc.GPT4oMiniInputPerMTok, rc.GPT4oMiniOutputPerMTok},
			{"Gemini Flash", rc.GeminiFlashInputPerMTok, rc.GeminiFlashOutputPerMTok},
			{"DeepSeek", rc.DeepSeekInputPerMTok, rc.DeepSeekOutputPerMTok},
		}
		for _, cand := range candidates {
			cost := inM*cand.in + outM*cand.out
			pct := 0.0
			if aiUsage.TotalCostUSD > 0 {
				pct = (cost - aiUsage.TotalCostUSD) / aiUsage.TotalCostUSD * 100
			}
			aiModels.Counterfactuals = append(aiModels.Counterfactuals, AICounterfactual{
				Label:       cand.label,
				CostUSD:     cost,
				VsActualPct: pct,
			})
		}
	}

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
		RateCard:            rc,
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
	c.AIModels = aiModels
	c.AIRegistry = aiRegistry
	c.AIOps = aiOps
	c.CacheEconomics = cacheEcon
	c.VideoEconomics = videoEcon
	c.Growth = growth
	c.RecipeQuality = recipeQuality
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

func (c *MetricCache) GetAIModels() *AIModelMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AIModels
}

func (c *MetricCache) GetAIRegistry() *queries.AIModelRegistry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AIRegistry
}

func (c *MetricCache) GetAIOps() *queries.AIOpsMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AIOps
}

func (c *MetricCache) GetCacheEconomics() *queries.CacheEconomics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CacheEconomics
}

func (c *MetricCache) GetVideoEconomics() *queries.VideoEconomics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.VideoEconomics
}

func (c *MetricCache) GetGrowth() *queries.GrowthMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Growth
}

func (c *MetricCache) GetRecipeQuality() *queries.RecipeQualityMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RecipeQuality
}

func (c *MetricCache) GetLastRefresh() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastRefresh
}
