package tasks

import (
	"context"
	"errors"
	"sort"
	"sync"

	"gorm.io/gorm"
)

type Repository interface {
	ListByUserID(ctx context.Context, userID uint) ([]Task, error)
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uint) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uint) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListByUserID(ctx context.Context, userID uint) ([]Task, error) {
	var items []Task
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id asc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *gormRepository) Create(ctx context.Context, task *Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id uint) (*Task, error) {
	var task Task
	err := r.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *gormRepository) Update(ctx context.Context, task *Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

type memoryRepository struct {
	mu     sync.RWMutex
	nextID uint
	tasks  map[uint]Task
}

func NewMemoryRepository() Repository {
	return &memoryRepository{
		nextID: 1,
		tasks:  make(map[uint]Task),
	}
}

func (r *memoryRepository) ListByUserID(_ context.Context, userID uint) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Task, 0)
	for _, task := range r.tasks {
		if task.UserID == userID {
			items = append(items, task)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

func (r *memoryRepository) Create(_ context.Context, task *Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.nextID
	r.nextID++
	r.tasks[task.ID] = *task
	return nil
}

func (r *memoryRepository) GetByID(_ context.Context, id uint) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	copy := task
	return &copy, nil
}

func (r *memoryRepository) Update(_ context.Context, task *Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[task.ID]; !ok {
		return ErrTaskNotFound
	}
	r.tasks[task.ID] = *task
	return nil
}

func (r *memoryRepository) Delete(_ context.Context, id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return ErrTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}
