package queries

import (
	"time"

	"gorm.io/gorm"
)

type UserMetrics struct {
	TotalUsers           int64            `json:"total_users"`
	NewUsersToday        int64            `json:"new_users_today"`
	NewUsersThisWeek     int64            `json:"new_users_this_week"`
	NewUsersThisMonth    int64            `json:"new_users_this_month"`
	UsersWithEmail       int64            `json:"users_with_email"`
	UsersWithPersonalization int64        `json:"users_with_personalization"`
	UnitSystemDistribution []LabelCount   `json:"unit_system_distribution"`
	DailyRegistrations   []DailyCount     `json:"daily_registrations"`
}

func GetUserMetrics(db *gorm.DB) (*UserMetrics, error) {
	m := &UserMetrics{}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	db.Table("users").Where("deleted_at IS NULL").Count(&m.TotalUsers)
	db.Table("users").Where("deleted_at IS NULL AND created_at >= ?", today).Count(&m.NewUsersToday)
	db.Table("users").Where("deleted_at IS NULL AND created_at >= ?", weekAgo).Count(&m.NewUsersThisWeek)
	db.Table("users").Where("deleted_at IS NULL AND created_at >= ?", monthAgo).Count(&m.NewUsersThisMonth)
	db.Table("users").Where("deleted_at IS NULL AND email IS NOT NULL AND email != ''").Count(&m.UsersWithEmail)
	db.Table("personalizations").Where("requirements != '' AND requirements IS NOT NULL").Count(&m.UsersWithPersonalization)

	db.Table("personalizations").
		Select("unit_system as label, COUNT(*) as count").
		Group("unit_system").
		Find(&m.UnitSystemDistribution)

	thirtyDaysAgo := today.AddDate(0, 0, -30)
	db.Table("users").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("deleted_at IS NULL AND created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Find(&m.DailyRegistrations)

	return m, nil
}
