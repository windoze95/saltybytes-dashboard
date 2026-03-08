package queries

import "time"

// Shared types used across query files.

type LabelCount struct {
	Label string `json:"label" gorm:"column:label"`
	Count int64  `json:"count" gorm:"column:count"`
}

type DailyCount struct {
	Date  time.Time `json:"date" gorm:"column:date"`
	Count int64     `json:"count" gorm:"column:count"`
}

type DailyLabelCount struct {
	Date  time.Time `json:"date" gorm:"column:date"`
	Label string    `json:"label" gorm:"column:label"`
	Count int64     `json:"count" gorm:"column:count"`
}

type TableSize struct {
	Name     string `json:"name" gorm:"column:name"`
	Rows     int64  `json:"rows" gorm:"column:rows"`
	SizeBytes int64 `json:"size_bytes" gorm:"column:size_bytes"`
}

type IndexSize struct {
	Name      string `json:"name" gorm:"column:name"`
	Table     string `json:"table" gorm:"column:table"`
	SizeBytes int64  `json:"size_bytes" gorm:"column:size_bytes"`
}

type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "fail", "warn"
	Count   int64  `json:"count"`
	Message string `json:"message"`
}

type TopEntry struct {
	Label string `json:"label" gorm:"column:label"`
	Count int64  `json:"count" gorm:"column:count"`
	Extra string `json:"extra,omitempty" gorm:"column:extra"`
}
