package carstats

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository abstracts stats persistence so the service can be unit-tested
// against an in-memory stub instead of a real database.
type Repository interface {
	LoadAll() ([]Stats, error)
	UpsertBatch(items []Stats) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) LoadAll() ([]Stats, error) {
	var items []Stats
	err := r.db.Find(&items).Error
	return items, err
}

func (r *gormRepository) UpsertBatch(items []Stats) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "car_id"}},
		UpdateAll: true,
	}).Create(&items).Error
}
