package queries

import (
	"gorm.io/gorm"
)

type FamilyMetrics struct {
	TotalFamilies         int64   `json:"total_families"`
	TotalMembers          int64   `json:"total_members"`
	AvgMembersPerFamily   float64 `json:"avg_members_per_family"`
	MembersWithDietaryProfile int64 `json:"members_with_dietary_profile"`
	DietaryProfileCoverage float64 `json:"dietary_profile_coverage"`
}

func GetFamilyMetrics(db *gorm.DB) (*FamilyMetrics, error) {
	m := &FamilyMetrics{}

	db.Table("families").Count(&m.TotalFamilies)
	db.Table("family_members").Count(&m.TotalMembers)
	db.Table("dietary_profiles").Count(&m.MembersWithDietaryProfile)

	if m.TotalMembers > 0 {
		m.AvgMembersPerFamily = float64(m.TotalMembers) / float64(m.TotalFamilies)
		m.DietaryProfileCoverage = float64(m.MembersWithDietaryProfile) / float64(m.TotalMembers) * 100
	}

	return m, nil
}
