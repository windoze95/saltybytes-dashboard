package queries

import (
	"time"

	"gorm.io/gorm"
)

// Shared types used across query files.

// safeScan extracts rows from a GORM query and scans into dest, gracefully
// handling query errors (e.g. missing columns) that would otherwise cause a
// nil-pointer panic via the standard .Row().Scan() path.
func safeScan(tx *gorm.DB, dest ...interface{}) {
	rows, err := tx.Rows()
	if err != nil || rows == nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(dest...)
	}
}

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
