package tasks

import (
	"context"
	"errors"
	"strings"

	"todo-backend/internal/users"
)

var (
	ErrInvalidTitle = errors.New("title must not be empty")
	ErrTaskNotFound = errors.New("task not found")
	ErrForbidden    = errors.New("task does not belong to user")
)

type UserReader interface {
	GetByID(ctx context.Context, id uint) (*users.User, error)
}

type Service struct {
	repo       Repository
	userReader UserReader
}

func NewService(repo Repository, userReader UserReader) *Service {
	return &Service{repo: repo, userReader: userReader}
}

func (s *Service) ListByUser(ctx context.Context, userID uint) ([]Task, error) {
	if _, err := s.userReader.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID uint, title string) (*Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrInvalidTitle
	}
	if _, err := s.userReader.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	task := &Task{Title: title, Done: false, UserID: userID}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Update(ctx context.Context, userID uint, taskID uint, title *string, done *bool) (*Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, ErrForbidden
	}

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return nil, ErrInvalidTitle
		}
		task.Title = trimmed
	}
	if done != nil {
		task.Done = *done
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Delete(ctx context.Context, userID uint, taskID uint) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, taskID)
}
