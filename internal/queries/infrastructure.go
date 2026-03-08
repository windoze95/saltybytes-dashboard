package queries

import (
	"gorm.io/gorm"
)

type InfrastructureMetrics struct {
	DatabaseSizeBytes int64       `json:"database_size_bytes"`
	DatabaseSizeMB    float64     `json:"database_size_mb"`
	TableSizes        []TableSize `json:"table_sizes"`
	IndexSizes        []IndexSize `json:"index_sizes"`
	ConnectionCount   int64       `json:"connection_count"`
	S3ImageCount      int64       `json:"s3_image_count"`
	S3EstimatedSizeMB float64     `json:"s3_estimated_size_mb"`
	S3EstimatedCost   float64     `json:"s3_estimated_cost"`
}

func GetInfrastructureMetrics(db *gorm.DB, s3CostPerGB float64) (*InfrastructureMetrics, error) {
	m := &InfrastructureMetrics{}

	// Total database size
	db.Raw("SELECT pg_database_size(current_database())").Row().Scan(&m.DatabaseSizeBytes)
	m.DatabaseSizeMB = float64(m.DatabaseSizeBytes) / (1024 * 1024)

	// Table sizes
	db.Raw(`SELECT
		relname as name,
		reltuples::bigint as rows,
		pg_total_relation_size(quote_ident(relname)) as size_bytes
		FROM pg_class
		WHERE relkind = 'r'
		AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		ORDER BY size_bytes DESC`).Find(&m.TableSizes)

	// Index sizes
	db.Raw(`SELECT
		i.relname as name,
		t.relname as table,
		pg_relation_size(i.oid) as size_bytes
		FROM pg_class i
		JOIN pg_index ix ON i.oid = ix.indexrelid
		JOIN pg_class t ON t.oid = ix.indrelid
		WHERE i.relkind = 'i'
		AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		ORDER BY size_bytes DESC`).Find(&m.IndexSizes)

	// Connection count
	db.Raw("SELECT COUNT(*) FROM pg_stat_activity WHERE datname = current_database()").Row().Scan(&m.ConnectionCount)

	// S3 image estimates
	db.Table("recipes").Where("deleted_at IS NULL AND image_url != '' AND image_url IS NOT NULL").Count(&m.S3ImageCount)
	avgImageSizeMB := 0.5 // ~500KB per image
	m.S3EstimatedSizeMB = float64(m.S3ImageCount) * avgImageSizeMB
	m.S3EstimatedCost = (m.S3EstimatedSizeMB / 1024) * s3CostPerGB

	return m, nil
}
