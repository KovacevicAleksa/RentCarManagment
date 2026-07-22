package notification

import "gorm.io/gorm"

// Repository abstracts notification persistence so the service can be
// unit-tested against an in-memory stub instead of a real database.
type Repository interface {
	Create(n *Notification) error
	FindAll(limit int) ([]Notification, error)
	MarkAllRead() error
	CountUnread() (int64, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(n *Notification) error {
	return r.db.Create(n).Error
}

func (r *gormRepository) FindAll(limit int) ([]Notification, error) {
	var items []Notification
	err := r.db.Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *gormRepository) MarkAllRead() error {
	return r.db.Model(&Notification{}).Where("read = ?", false).Update("read", true).Error
}

func (r *gormRepository) CountUnread() (int64, error) {
	var n int64
	err := r.db.Model(&Notification{}).Where("read = ?", false).Count(&n).Error
	return n, err
}
