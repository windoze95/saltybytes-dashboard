package queries

import (
	"time"

	"gorm.io/gorm"
)

// AIModelOptionRow mirrors a row of the API's ai_model_options registry table.
type AIModelOptionRow struct {
	ID                 uint       `json:"id" gorm:"column:id"`
	Provider           string     `json:"provider" gorm:"column:provider"`
	ModelID            string     `json:"model_id" gorm:"column:model_id"`
	Label              string     `json:"label" gorm:"column:label"`
	BaseURL            string     `json:"base_url" gorm:"column:base_url"`
	InputPricePerMTok  float64    `json:"input_price_per_mtok" gorm:"column:input_price_per_mtok"`
	OutputPricePerMTok float64    `json:"output_price_per_mtok" gorm:"column:output_price_per_mtok"`
	Enabled            bool       `json:"enabled" gorm:"column:enabled"`
	Validated          bool       `json:"validated" gorm:"column:validated"`
	ValidationError    string     `json:"validation_error" gorm:"column:validation_error"`
	LastValidatedAt    *time.Time `json:"last_validated_at" gorm:"column:last_validated_at"`
}

// AIActiveSpec mirrors the single ai_config row: the currently-active light tier.
type AIActiveSpec struct {
	ActiveProvider string `json:"active_provider" gorm:"column:active_provider"`
	ActiveModel    string `json:"active_model" gorm:"column:active_model"`
	ActiveBaseURL  string `json:"active_base_url" gorm:"column:active_base_url"`
}

// AIModelRegistry is the light-tier model registry plus the active selection,
// read directly (read-only) from the API's tables.
type AIModelRegistry struct {
	Options []AIModelOptionRow `json:"options"`
	Active  AIActiveSpec       `json:"active"`
}

// GetAIModelRegistry reads ai_model_options + ai_config. Drift-tolerant: if the
// tables don't exist yet (the API hasn't deployed the registry migration) it
// returns an empty registry rather than erroring.
func GetAIModelRegistry(db *gorm.DB) (*AIModelRegistry, error) {
	reg := &AIModelRegistry{Options: []AIModelOptionRow{}}

	if hasColumn(db, "ai_model_options", "id") {
		db.Table("ai_model_options").
			Select("id, provider, model_id, label, base_url, input_price_per_mtok, output_price_per_mtok, enabled, validated, validation_error, last_validated_at").
			Order("provider, model_id").
			Find(&reg.Options)
	}

	if hasColumn(db, "ai_config", "active_provider") {
		safeScan(
			db.Table("ai_config").Select("active_provider, active_model, active_base_url").Order("id").Limit(1),
			&reg.Active.ActiveProvider, &reg.Active.ActiveModel, &reg.Active.ActiveBaseURL,
		)
	}

	return reg, nil
}
