package tasks

import "time"

type Task struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Done      bool      `json:"done" gorm:"not null;default:false"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	AIGeneratedSummary string `json:"ai_summary"`
	AIPriority         string `json:"ai_priority"`
	AITags             string `json:"ai_tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
