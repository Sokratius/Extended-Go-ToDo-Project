package users

import (
	"context"
	"errors"
	"sync"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

type memoryRepository struct {
	mu     sync.RWMutex
	nextID uint
	users  map[uint]User
}

func NewMemoryRepository() Repository {
	return &memoryRepository{
		nextID: 1,
		users:  make(map[uint]User),
	}
}

func (r *memoryRepository) Create(_ context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.users {
		if existing.Username == user.Username {
			return ErrUsernameTaken
		}
	}

	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = *user
	return nil
}

func (r *memoryRepository) GetByUsername(_ context.Context, username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			copy := user
			return &copy, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *memoryRepository) GetByID(_ context.Context, id uint) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	copy := user
	return &copy, nil
}
