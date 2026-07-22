package carstats

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository abstracts stats persistence so the service can be unit-tested
// against an in-memory stub instead of a real database.
type Repository interface {
	LoadAll() ([]Stats, error)
	UpsertBatch(items []Stats) error
	// AppendEvents stores newly observed fault-episode events.
	AppendEvents(events []Event) error
	// LoadEventsSince returns all events at or after the given time, used to
	// rehydrate the reliability window on startup.
	LoadEventsSince(since time.Time) ([]Event, error)
	// LoadEventsByCarSince returns one car's events at or after the given time,
	// used to build the per-car fault history for the dashboard charts.
	LoadEventsByCarSince(carID string, since time.Time) ([]Event, error)
	// DeleteEventsBefore prunes events older than the given time so the table
	// stays bounded to roughly the reliability window.
	DeleteEventsBefore(cutoff time.Time) error
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

func (r *gormRepository) AppendEvents(events []Event) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.Create(&events).Error
}

func (r *gormRepository) LoadEventsSince(since time.Time) ([]Event, error) {
	var events []Event
	err := r.db.Where("occurred_at >= ?", since).Find(&events).Error
	return events, err
}

func (r *gormRepository) LoadEventsByCarSince(carID string, since time.Time) ([]Event, error) {
	var events []Event
	err := r.db.Where("car_id = ? AND occurred_at >= ?", carID, since).
		Order("occurred_at ASC").
		Find(&events).Error
	return events, err
}

func (r *gormRepository) DeleteEventsBefore(cutoff time.Time) error {
	return r.db.Where("occurred_at < ?", cutoff).Delete(&Event{}).Error
}
