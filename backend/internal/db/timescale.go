package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

func EnableTimescaleDB(db *gorm.DB) error {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE").Error; err != nil {
		return fmt.Errorf("failed to enable timescaledb: %w", err)
	}
	log.Println("TimescaleDB enabled")
	return nil
}

func ConvertToHypertable(db *gorm.DB, tableName string, timeColumn string) error {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM timescaledb_information.hypertables 
			WHERE hypertable_name = ?
		)
	`
	if err := db.Raw(query, tableName).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check hypertable: %w", err)
	}

	if exists {
		log.Printf("Table %s already hypertable", tableName)
		return nil
	}

	sql := fmt.Sprintf("SELECT create_hypertable('%s', '%s', if_not_exists => TRUE)", tableName, timeColumn)
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to create hypertable: %w", err)
	}

	log.Printf("Table %s converted to hypertable", tableName)
	return nil
}

func CreateCompressionPolicy(db *gorm.DB, tableName string, olderThan string) error {
	sql := fmt.Sprintf(`
		ALTER TABLE %s SET (
			timescaledb.compress,
			timescaledb.compress_segmentby = 'car_id'
		)
	`, tableName)
	
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to enable compression: %w", err)
	}

	sql = fmt.Sprintf(`
		SELECT add_compression_policy('%s', INTERVAL '%s')
	`, tableName, olderThan)
	
	if err := db.Exec(sql).Error; err != nil {
		log.Printf("Compression policy: %v", err)
	} else {
		log.Printf("Compression policy created for %s", tableName)
	}

	return nil
}

func CreateRetentionPolicy(db *gorm.DB, tableName string, retainFor string) error {
	sql := fmt.Sprintf(`
		SELECT add_retention_policy('%s', INTERVAL '%s')
	`, tableName, retainFor)
	
	if err := db.Exec(sql).Error; err != nil {
		log.Printf("Retention policy: %v", err)
	} else {
		log.Printf("Retention policy created for %s", tableName)
	}

	return nil
}