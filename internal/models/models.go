package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── User ───────────────────────────────────────────────────────────────────

type User struct {
	gorm.Model
	Username        string           `gorm:"uniqueIndex" json:"username"`
	FirstName       *string          `gorm:"default:null" json:"first_name"`
	Email           *string          `gorm:"uniqueIndex;default:null" json:"email"`
	Subscription    *Subscription    `gorm:"foreignKey:UserID" json:"subscription,omitempty"`
	Settings        *UserSettings    `gorm:"foreignKey:UserID" json:"settings,omitempty"`
	Personalization *Personalization `gorm:"foreignKey:UserID" json:"personalization,omitempty"`
}

func (User) TableName() string { return "users" }

// ─── Subscription ───────────────────────────────────────────────────────────

type SubscriptionTier string

const (
	TierFree    SubscriptionTier = "free"
	TierPremium SubscriptionTier = "premium"
)

type Subscription struct {
	gorm.Model
	UserID               uint             `gorm:"uniqueIndex;not null" json:"user_id"`
	Tier                 SubscriptionTier `gorm:"type:text;default:'free'" json:"tier"`
	ExpiresAt            *time.Time       `json:"expires_at"`
	AllergenAnalysesUsed int              `gorm:"default:0" json:"allergen_analyses_used"`
	WebSearchesUsed      int              `gorm:"default:0" json:"web_searches_used"`
	AIGenerationsUsed    int              `gorm:"default:0" json:"ai_generations_used"`
	MonthlyResetAt       time.Time        `json:"monthly_reset_at"`
}

func (Subscription) TableName() string { return "subscriptions" }

// ─── UserSettings ───────────────────────────────────────────────────────────

type UserSettings struct {
	gorm.Model
	UserID         uint `gorm:"uniqueIndex" json:"user_id"`
	KeepScreenAwake bool `gorm:"default:true" json:"keep_screen_awake"`
}

func (UserSettings) TableName() string { return "user_settings" }

// ─── Personalization ────────────────────────────────────────────────────────

type Personalization struct {
	gorm.Model
	UserID       uint      `gorm:"uniqueIndex" json:"user_id"`
	UnitSystem   string    `gorm:"type:text;default:'us_customary'" json:"unit_system"`
	Requirements string    `json:"requirements"`
	UID          uuid.UUID `gorm:"type:uuid" json:"uid"`
}

func (Personalization) TableName() string { return "personalizations" }

// ─── Recipe ─────────────────────────────────────────────────────────────────

type RecipeType string

const (
	RecipeTypeChat         RecipeType = "chat"
	RecipeTypeRegenChat    RecipeType = "regen_chat"
	RecipeTypeFork         RecipeType = "fork"
	RecipeTypeImportVision RecipeType = "import_vision"
	RecipeTypeImportLink   RecipeType = "import_link"
	RecipeTypeImportText   RecipeType = "import_text"
	RecipeTypeManualEntry  RecipeType = "user_input"
)

type Recipe struct {
	gorm.Model
	Title            string      `json:"title"`
	Ingredients      Ingredients `gorm:"type:jsonb" json:"ingredients"`
	Instructions     StringArray `gorm:"type:text[]" json:"instructions"`
	CookTime         int         `json:"cook_time"`
	ImagePrompt      string      `json:"image_prompt"`
	ImageURL         string      `json:"image_url"`
	Portions         int         `json:"portions,omitempty"`
	PortionSize      string      `json:"portion_size,omitempty"`
	SourceURL        string      `json:"source_url,omitempty"`
	UnitSystem       string      `json:"unit_system,omitempty"`
	Status           string      `gorm:"default:'ready'" json:"status"`
	CreatedByID      uint        `json:"created_by_id"`
	UserEdited       bool        `gorm:"default:false" json:"user_edited"`
	ForkedFromID     *uint       `gorm:"index" json:"forked_from_id"`
	TreeID           *uint       `gorm:"index" json:"tree_id"`
	OriginalImageURL string      `json:"original_image_url"`
	CanonicalID      *uint       `gorm:"index" json:"canonical_id"`
	HasDiverged      bool        `gorm:"default:false" json:"has_diverged"`
}

func (Recipe) TableName() string { return "recipes" }

// ─── Tag ────────────────────────────────────────────────────────────────────

type Tag struct {
	gorm.Model
	Hashtag string `gorm:"index;unique" json:"hashtag"`
}

func (Tag) TableName() string { return "tags" }

// ─── RecipeTree ─────────────────────────────────────────────────────────────

type RecipeTree struct {
	gorm.Model
	RecipeID   uint  `gorm:"uniqueIndex;not null" json:"recipe_id"`
	RootNodeID *uint `gorm:"index" json:"root_node_id"`
}

func (RecipeTree) TableName() string { return "recipe_trees" }

// ─── RecipeNode ─────────────────────────────────────────────────────────────

type RecipeNode struct {
	gorm.Model
	TreeID      uint       `gorm:"index" json:"tree_id"`
	ParentID    *uint      `gorm:"index" json:"parent_id"`
	Prompt      string     `json:"prompt"`
	Summary     string     `json:"summary"`
	Type        RecipeType `gorm:"type:text" json:"type"`
	BranchName  string     `gorm:"default:'original'" json:"branch_name"`
	IsEphemeral bool       `gorm:"default:false" json:"is_ephemeral"`
	CreatedByID uint       `gorm:"index" json:"created_by_id"`
	IsActive    bool       `gorm:"default:false" json:"is_active"`
}

func (RecipeNode) TableName() string { return "recipe_nodes" }

// ─── Canonical Recipe ───────────────────────────────────────────────────────

type ExtractionMethod string

const (
	ExtractionJSONLD          ExtractionMethod = "json_ld"
	ExtractionHaiku           ExtractionMethod = "haiku"
	ExtractionFirecrawlJSONLD ExtractionMethod = "firecrawl_json_ld"
	ExtractionFirecrawlHaiku  ExtractionMethod = "firecrawl_haiku"
)

type CanonicalRecipe struct {
	gorm.Model
	NormalizedURL    string           `gorm:"uniqueIndex;size:2048" json:"normalized_url"`
	OriginalURL      string           `json:"original_url"`
	ExtractionMethod ExtractionMethod `gorm:"type:text" json:"extraction_method"`
	HitCount         int              `gorm:"default:0" json:"hit_count"`
	LastAccessedAt   time.Time        `json:"last_accessed_at"`
	FetchedAt        time.Time        `json:"fetched_at"`
}

func (CanonicalRecipe) TableName() string { return "canonical_recipes" }

// ─── Search Cache ───────────────────────────────────────────────────────────

type SearchCache struct {
	gorm.Model
	NormalizedQuery string    `gorm:"uniqueIndex;size:512" json:"normalized_query"`
	ResultCount     int       `json:"result_count"`
	HitCount        int       `gorm:"default:0" json:"hit_count"`
	LastAccessedAt  time.Time `json:"last_accessed_at"`
	FetchedAt       time.Time `json:"fetched_at"`
}

func (SearchCache) TableName() string { return "search_caches" }

// ─── Allergen Analysis ──────────────────────────────────────────────────────

type AllergenAnalysis struct {
	gorm.Model
	RecipeID         uint    `gorm:"index" json:"recipe_id"`
	NodeID           *uint   `gorm:"index" json:"node_id"`
	ContainsNuts     bool    `json:"contains_nuts"`
	ContainsDairy    bool    `json:"contains_dairy"`
	ContainsGluten   bool    `json:"contains_gluten"`
	ContainsSoy      bool    `json:"contains_soy"`
	ContainsSeedOils bool    `json:"contains_seed_oils"`
	ContainsShellfish bool   `json:"contains_shellfish"`
	ContainsEggs     bool    `json:"contains_eggs"`
	Confidence       float64 `json:"confidence"`
	RequiresReview   bool    `json:"requires_review"`
	IsPremium        bool    `json:"is_premium"`
}

func (AllergenAnalysis) TableName() string { return "allergen_analyses" }

// ─── Family ─────────────────────────────────────────────────────────────────

type Family struct {
	gorm.Model
	Name    string         `json:"name"`
	OwnerID uint           `json:"owner_id"`
	Members []FamilyMember `gorm:"foreignKey:FamilyID" json:"members"`
}

func (Family) TableName() string { return "families" }

type FamilyMember struct {
	gorm.Model
	FamilyID       uint            `gorm:"index" json:"family_id"`
	Name           string          `json:"name"`
	Relationship   string          `json:"relationship"`
	UserID         *uint           `json:"user_id"`
	DietaryProfile *DietaryProfile `gorm:"foreignKey:MemberID" json:"dietary_profile"`
}

func (FamilyMember) TableName() string { return "family_members" }

type DietaryProfile struct {
	gorm.Model
	MemberID uint `gorm:"uniqueIndex" json:"member_id"`
}

func (DietaryProfile) TableName() string { return "dietary_profiles" }

// ─── JSON/Array helpers ─────────────────────────────────────────────────────

type Ingredient struct {
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	Amount       float64 `json:"amount"`
	MetricUnit   string  `json:"metric_unit,omitempty"`
	MetricAmount float64 `json:"metric_amount,omitempty"`
	OriginalText string  `json:"original_text,omitempty"`
}

type Ingredients []Ingredient

func (i *Ingredients) Scan(value interface{}) error {
	if value == nil {
		*i = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Ingredients: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, i)
}

func (i Ingredients) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	return json.Marshal(i)
}

type StringArray []string

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringArray: expected []byte, got %T", value)
	}
	// PostgreSQL text[] comes as {val1,val2,...}
	str := string(bytes)
	if str == "{}" || str == "" {
		*s = nil
		return nil
	}
	// Use json if it looks like JSON array
	if str[0] == '[' {
		return json.Unmarshal(bytes, s)
	}
	// Otherwise parse PG array format
	return json.Unmarshal(bytes, s)
}

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}
