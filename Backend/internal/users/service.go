package users

import (
	"context"
	"errors"
	"strings"

	"todo-backend/pkg/utils"
)

var (
	ErrInvalidUsername    = errors.New("username must not be empty")
	ErrInvalidPassword    = errors.New("password must not be empty")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, username, password string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, ErrInvalidUsername
	}
	if strings.TrimSpace(password) == "" {
		return nil, ErrInvalidPassword
	}

	if _, err := s.repo.GetByUsername(ctx, username); err == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username: username,
		Password: hashedPassword,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*User, error) {
	user, err := s.repo.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := utils.CheckPasswordHash(password, user.Password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*User, error) {
	return s.repo.GetByID(ctx, id)
}
