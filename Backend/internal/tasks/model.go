package tasks

import "time"

type Task struct {
	ID        		uint      `json:"id" gorm:"primaryKey"`
	Title     		string    `json:"title" gorm:"not null"`
	Done      		bool      `json:"done" gorm:"not null;default:false"`
	Description		string    `json:"description"`
	UserID    		uint      `json:"user_id" gorm:"index;not null"`
	AIPriority         string    `json:"ai_priority" gorm:"type:varchar(20);default:'Medium'"` 
	AITags             string    `json:"ai_tags"` 
	AIGeneratedSummary string    `json:"ai_summary" gorm:"type:text"` 
	EstimatedTimeMin   int       `json:"estimated_time_min"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
