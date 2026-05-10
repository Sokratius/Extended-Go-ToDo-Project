package notifications

import (
	"context"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, notification *Notification) error
	GetUnreadByUserID(ctx context.Context, userID uint) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID uint, userID uint) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *gormRepository) GetUnreadByUserID(ctx context.Context, userID uint) ([]Notification, error) {
	var items []Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at desc").
		Find(&items).Error
	return items, err
}

func (r *gormRepository) MarkAsRead(ctx context.Context, notificationID uint, userID uint) error {
	result := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	return result.Error
}