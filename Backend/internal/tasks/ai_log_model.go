package tasks

import "time"

type AILog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	TaskID       uint      `json:"task_id" gorm:"index"` 
	PromptType   string    `json:"prompt_type"` 
	RequestText  string    `json:"request_text" gorm:"type:text"` 
	ResponseText string    `json:"response_text" gorm:"type:text"` 
	IsSuccessful bool      `json:"is_successful" gorm:"default:true"` 
	CreatedAt    time.Time `json:"created_at"`
}