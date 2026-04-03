package tasks

import "context"

type ServiceInterface interface {
	ListByUser(ctx context.Context, userID uint) ([]Task, error)
	Create(ctx context.Context, userID uint, title string) (*Task, error)
	Update(ctx context.Context, userID, taskID uint, title *string, done *bool) (*Task, error)
	Delete(ctx context.Context, userID, taskID uint) error
}
