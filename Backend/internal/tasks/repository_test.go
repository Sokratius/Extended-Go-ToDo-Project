package tasks

import (
	"context"
	"testing"
)

func newRepo() Repository {
	return NewMemoryRepository()
}

func TestCreate(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	task := &Task{Title: "Buy milk", UserID: 1}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.ID == 0 {
		t.Error("expected ID to be set after Create")
	}
}

func TestGetByID_Found(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	task := &Task{Title: "Read book", UserID: 2}
	_ = repo.Create(ctx, task)

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != task.Title {
		t.Errorf("expected title %q, got %q", task.Title, got.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestListByUserID(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	_ = repo.Create(ctx, &Task{Title: "Task 1", UserID: 1})
	_ = repo.Create(ctx, &Task{Title: "Task 2", UserID: 1})
	_ = repo.Create(ctx, &Task{Title: "Task 3", UserID: 2})

	items, err := repo.ListByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 tasks for user 1, got %d", len(items))
	}
}

func TestListByUserID_OrderedByID(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	_ = repo.Create(ctx, &Task{Title: "First", UserID: 1})
	_ = repo.Create(ctx, &Task{Title: "Second", UserID: 1})
	_ = repo.Create(ctx, &Task{Title: "Third", UserID: 1})

	items, _ := repo.ListByUserID(ctx, 1)
	for i := 1; i < len(items); i++ {
		if items[i].ID < items[i-1].ID {
			t.Errorf("tasks not ordered by ID: %d before %d", items[i-1].ID, items[i].ID)
		}
	}
}

func TestUpdate(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	task := &Task{Title: "Old title", UserID: 1}
	_ = repo.Create(ctx, task)

	task.Title = "New title"
	task.Done = true
	if err := repo.Update(ctx, task); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, task.ID)
	if got.Title != "New title" {
		t.Errorf("expected title %q, got %q", "New title", got.Title)
	}
	if !got.Done {
		t.Error("expected Done to be true")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	err := repo.Update(ctx, &Task{ID: 999, Title: "Ghost"})
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	task := &Task{Title: "To delete", UserID: 1}
	_ = repo.Create(ctx, task)

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, task.ID)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound after delete, got %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	err := repo.Delete(ctx, 999)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}
