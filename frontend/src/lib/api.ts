const BASE = '/api'

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { credentials: 'include' })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

async function postJSON<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

async function putJSON<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

export const api = {
  login: (password: string) => postJSON<{ status: string }>('/auth/login', { password }),
  logout: () => postJSON<{ status: string }>('/auth/logout'),
  checkAuth: () => fetchJSON<{ authenticated: boolean }>('/auth/check'),

  overview: () => fetchJSON<OverviewMetrics>('/overview'),
  users: () => fetchJSON<UserMetrics>('/users'),
  recipes: () => fetchJSON<RecipeMetrics>('/recipes'),
  canonical: () => fetchJSON<CanonicalMetrics>('/canonical'),
  searchCache: () => fetchJSON<SearchCacheMetrics>('/search-cache'),
  subscriptions: () => fetchJSON<SubscriptionMetrics>('/subscriptions'),
  allergens: () => fetchJSON<AllergenMetrics>('/allergens'),
  families: () => fetchJSON<FamilyMetrics>('/families'),
  infrastructure: () => fetchJSON<InfrastructureMetrics>('/infrastructure'),
  healthChecks: () => fetchJSON<HealthCheck[]>('/health-checks'),
  costCenter: () => fetchJSON<CostCenterMetrics>('/cost-center'),
  rateCard: () => fetchJSON<RateCard>('/rate-card'),
  updateRateCard: (card: RateCard) => putJSON<RateCard>('/rate-card/update', card),
  refresh: () => postJSON<{ status: string }>('/refresh'),
}

// Types

export interface LabelCount {
  label: string
  count: number
}

export interface DailyCount {
  date: string
  count: number
}

export interface DailyLabelCount {
  date: string
  label: string
  count: number
}

export interface TopEntry {
  label: string
  count: number
  extra?: string
}

export interface HealthCheck {
  name: string
  status: 'pass' | 'fail' | 'warn'
  count: number
  message: string
}

export interface OverviewMetrics {
  total_users: number
  total_recipes: number
  total_canonicals: number
  today_estimated_cost: number
  month_to_date_cost: number
  month_projection: number
  search_cache_hit_rate: number
  firecrawl_credits_used: number
  firecrawl_credits_left: number
  health_checks_passing: number
  health_checks_total: number
  recipes_created_week: DailyCount[]
}

export interface UserMetrics {
  total_users: number
  new_users_today: number
  new_users_this_week: number
  new_users_this_month: number
  users_with_email: number
  users_with_personalization: number
  unit_system_distribution: LabelCount[]
  daily_registrations: DailyCount[]
}

export interface RecipeMetrics {
  total_recipes: number
  recipes_today: number
  recipes_this_week: number
  recipes_with_images: number
  recipes_with_embeddings: number
  recipes_with_source_url: number
  embedding_coverage: number
  image_coverage: number
  user_edited_rate: number
  user_edited_count: number
  avg_ingredients_per_recipe: number
  avg_cook_time: number
  avg_portions: number
  image_regen_count: number
  deleted_today: number
  daily_creation: DailyLabelCount[]
  node_type_distribution: LabelCount[]
  top_hashtags: TopEntry[]
  recipes_per_user: LabelCount[]
  import_breakdown: LabelCount[]
  canonical_linked: number
  canonical_non_diverged: number
  canonical_diverged: number
  divergence_rate: number
  max_fork_depth: number
  avg_fork_depth: number
  fork_count: number
  total_trees: number
  ephemeral_nodes: number
}

export interface CanonicalMetrics {
  total_entries: number
  new_today: number
  avg_hit_count: number
  zero_hit_entries: number
  extraction_method_distribution: LabelCount[]
  top_by_hits: TopEntry[]
  entries_approaching_ttl: number
  hot_entries_eligible: number
  stale_entries_90d: number
  firecrawl_calls_this_month: number
  firecrawl_total_calls: number
  total_ai_extractions_saved: number
  daily_new: DailyCount[]
  top_domains: TopEntry[]
}

export interface SearchCacheMetrics {
  total_cached_queries: number
  avg_hit_count: number
  zero_hit_entries: number
  entries_with_embeddings: number
  embedding_coverage: number
  avg_results_per_query: number
  top_queries: TopEntry[]
  entries_approaching_ttl: number
  stale_entries: number
  hot_queries_eligible: number
  stale_entries_30d: number
  daily_volume: DailyCount[]
}

export interface SubscriptionMetrics {
  tier_distribution: LabelCount[]
  avg_allergen_used: number
  avg_searches_used: number
  avg_ai_generations_used: number
  users_near_allergen_limit: number
  users_near_search_limit: number
  users_near_ai_gen_limit: number
  users_at_limit: UserAtLimit[]
  allergen_distribution: LabelCount[]
  search_distribution: LabelCount[]
  ai_gen_distribution: LabelCount[]
}

export interface UserAtLimit {
  user_id: number
  username: string
  tier: string
  allergen_analyses_used: number
  web_searches_used: number
  ai_generations_used: number
}

export interface AllergenMetrics {
  total_analyses: number
  analyses_today: number
  avg_confidence: number
  requires_review_rate: number
  requires_review_count: number
  premium_count: number
  free_count: number
  allergen_flags: LabelCount[]
  daily_volume: DailyCount[]
  confidence_distribution: LabelCount[]
}

export interface FamilyMetrics {
  total_families: number
  total_members: number
  avg_members_per_family: number
  members_with_dietary_profile: number
  dietary_profile_coverage: number
}

export interface InfrastructureMetrics {
  database_size_bytes: number
  database_size_mb: number
  table_sizes: TableSize[]
  index_sizes: IndexSize[]
  connection_count: number
  s3_image_count: number
  s3_estimated_size_mb: number
  s3_estimated_cost: number
}

export interface TableSize {
  name: string
  rows: number
  size_bytes: number
}

export interface IndexSize {
  name: string
  table: string
  size_bytes: number
}

export interface CostCenterMetrics {
  daily_cost_by_provider: DailyLabelCount[]
  cost_by_feature: LabelCount[]
  cost_per_user: number
  cost_per_recipe: number
  search_cache_savings: number
  canonical_cache_savings: number
  allergen_cache_savings: number
  total_savings: number
  firecrawl_credits_used: number
  firecrawl_credits_max: number
  monthly_fixed_costs: number
  rate_card: RateCard
}

export interface RateCard {
  anthropic_sonnet_input_per_mtok: number
  anthropic_sonnet_output_per_mtok: number
  anthropic_haiku_input_per_mtok: number
  anthropic_haiku_output_per_mtok: number
  openai_dalle_per_image: number
  openai_whisper_per_minute: number
  openai_embedding_per_mtok: number
  brave_monthly_plan: number
  firecrawl_monthly_credits: number
  firecrawl_credits_per_scrape: number
  aws_rds_monthly: number
  aws_ecs_monthly: number
  aws_s3_per_gb: number
}
