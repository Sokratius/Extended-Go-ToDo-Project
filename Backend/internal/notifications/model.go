package notifications

import "time"

type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"` 
	TaskID    uint      `json:"task_id"`                       
	Message   string    `json:"message" gorm:"type:text;not null"` 
	Type      string    `json:"type" gorm:"type:varchar(50);not null"` 
	IsRead    bool      `json:"is_read" gorm:"default:false;index"` 
	CreatedAt time.Time `json:"created_at"`
}