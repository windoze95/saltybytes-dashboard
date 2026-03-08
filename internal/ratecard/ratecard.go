package ratecard

import (
	"encoding/json"
	"os"
	"sync"
)

type RateCard struct {
	mu sync.RWMutex

	// Anthropic
	AnthropicSonnetInputPerMTok  float64 `json:"anthropic_sonnet_input_per_mtok"`
	AnthropicSonnetOutputPerMTok float64 `json:"anthropic_sonnet_output_per_mtok"`
	AnthropicHaikuInputPerMTok   float64 `json:"anthropic_haiku_input_per_mtok"`
	AnthropicHaikuOutputPerMTok  float64 `json:"anthropic_haiku_output_per_mtok"`

	// OpenAI
	OpenAIDallePerImage    float64 `json:"openai_dalle_per_image"`
	OpenAIWhisperPerMinute float64 `json:"openai_whisper_per_minute"`
	OpenAIEmbeddingPerMTok float64 `json:"openai_embedding_per_mtok"`

	// Search & Scraping
	BraveMonthlyPlan          float64 `json:"brave_monthly_plan"`
	FirecrawlMonthlyCredits   int     `json:"firecrawl_monthly_credits"`
	FirecrawlCreditsPerScrape int     `json:"firecrawl_credits_per_scrape"`

	// AWS
	AWSRDSMonthly float64 `json:"aws_rds_monthly"`
	AWSECSMonthly float64 `json:"aws_ecs_monthly"`
	AWSS3PerGB    float64 `json:"aws_s3_per_gb"`
}

func Default() *RateCard {
	return &RateCard{
		AnthropicSonnetInputPerMTok:  3.00,
		AnthropicSonnetOutputPerMTok: 15.00,
		AnthropicHaikuInputPerMTok:   0.80,
		AnthropicHaikuOutputPerMTok:  4.00,
		OpenAIDallePerImage:          0.04,
		OpenAIWhisperPerMinute:       0.006,
		OpenAIEmbeddingPerMTok:       0.02,
		BraveMonthlyPlan:             0.00,
		FirecrawlMonthlyCredits:      500,
		FirecrawlCreditsPerScrape:    1,
		AWSRDSMonthly:                0.00,
		AWSECSMonthly:                0.00,
		AWSS3PerGB:                   0.023,
	}
}

func (rc *RateCard) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // use defaults
		}
		return err
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return json.Unmarshal(data, rc)
}

func (rc *RateCard) Save(path string) error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	data, err := json.MarshalIndent(rc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (rc *RateCard) Get() RateCard {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return *rc
}

func (rc *RateCard) Update(updated RateCard) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.AnthropicSonnetInputPerMTok = updated.AnthropicSonnetInputPerMTok
	rc.AnthropicSonnetOutputPerMTok = updated.AnthropicSonnetOutputPerMTok
	rc.AnthropicHaikuInputPerMTok = updated.AnthropicHaikuInputPerMTok
	rc.AnthropicHaikuOutputPerMTok = updated.AnthropicHaikuOutputPerMTok
	rc.OpenAIDallePerImage = updated.OpenAIDallePerImage
	rc.OpenAIWhisperPerMinute = updated.OpenAIWhisperPerMinute
	rc.OpenAIEmbeddingPerMTok = updated.OpenAIEmbeddingPerMTok
	rc.BraveMonthlyPlan = updated.BraveMonthlyPlan
	rc.FirecrawlMonthlyCredits = updated.FirecrawlMonthlyCredits
	rc.FirecrawlCreditsPerScrape = updated.FirecrawlCreditsPerScrape
	rc.AWSRDSMonthly = updated.AWSRDSMonthly
	rc.AWSECSMonthly = updated.AWSECSMonthly
	rc.AWSS3PerGB = updated.AWSS3PerGB
}
