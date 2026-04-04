package tasks

import (
	"context"
	"testing"

	"todo-backend/internal/users"
)

// stubUserReader реализует UserReader для тестов.
type stubUserReader struct {
	users map[uint]*users.User
}

func (s *stubUserReader) GetByID(_ context.Context, id uint) (*users.User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, users.ErrUserNotFound
	}
	return u, nil
}

func newService() (*Service, *stubUserReader) {
	reader := &stubUserReader{
		users: map[uint]*users.User{
			1: {ID: 1, Username: "alice"},
			2: {ID: 2, Username: "bob"},
		},
	}
	svc := NewService(NewMemoryRepository(), reader)
	return svc, reader
}

// --- ListByUser ---

func TestService_ListByUser_Empty(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	tasks, err := svc.ListByUser(ctx, 1)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestService_ListByUser_ReturnOnlyOwn(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	svc.Create(ctx, 1, "Task A")
	svc.Create(ctx, 1, "Task B")
	svc.Create(ctx, 2, "Task C")

	tasks, err := svc.ListByUser(ctx, 1)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for user 1, got %d", len(tasks))
	}
}

func TestService_ListByUser_UserNotFound(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	_, err := svc.ListByUser(ctx, 999)
	if err != users.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// --- Create ---

func TestService_Create_Success(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, err := svc.Create(ctx, 1, "Buy milk")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.ID == 0 {
		t.Error("expected ID to be set")
	}
	if task.Title != "Buy milk" {
		t.Errorf("expected title %q, got %q", "Buy milk", task.Title)
	}
	if task.UserID != 1 {
		t.Errorf("expected userID 1, got %d", task.UserID)
	}
	if task.Done {
		t.Error("new task should not be done")
	}
}

func TestService_Create_TrimsTitle(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, err := svc.Create(ctx, 1, "  Read book  ")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.Title != "Read book" {
		t.Errorf("expected trimmed title, got %q", task.Title)
	}
}

func TestService_Create_EmptyTitle(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	_, err := svc.Create(ctx, 1, "   ")
	if err != ErrInvalidTitle {
		t.Errorf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestService_Create_UserNotFound(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	_, err := svc.Create(ctx, 999, "Task")
	if err != users.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// --- Update ---

func TestService_Update_Title(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, _ := svc.Create(ctx, 1, "Old title")
	newTitle := "New title"

	updated, err := svc.Update(ctx, 1, task.ID, &newTitle, nil)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "New title" {
		t.Errorf("expected %q, got %q", "New title", updated.Title)
	}
}

func TestService_Update_Done(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, _ := svc.Create(ctx, 1, "Task")
	done := true

	updated, err := svc.Update(ctx, 1, task.ID, nil, &done)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !updated.Done {
		t.Error("expected Done to be true")
	}
}

func TestService_Update_BothFields(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, _ := svc.Create(ctx, 1, "Task")
	newTitle := "Updated"
	done := true

	updated, err := svc.Update(ctx, 1, task.ID, &newTitle, &done)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Updated" || !updated.Done {
		t.Errorf("unexpected state: title=%q done=%v", updated.Title, updated.Done)
	}
}

func TestService_Update_EmptyTitle(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, _ := svc.Create(ctx, 1, "Task")
	empty := "   "

	_, err := svc.Update(ctx, 1, task.ID, &empty, nil)
	if err != ErrInvalidTitle {
		t.Errorf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestService_Update_TaskNotFound(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	title := "x"
	_, err := svc.Update(ctx, 1, 999, &title, nil)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestService_Update_Forbidden(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, _ := svc.Create(ctx, 1, "Task")
	title := "Hacked"

	_, err := svc.Update(ctx, 2, task.ID, &title, nil)
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- Delete ---

func TestService_Delete_Success(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, _ := svc.Create(ctx, 1, "Task")

	if err := svc.Delete(ctx, 1, task.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	tasks, _ := svc.ListByUser(ctx, 1)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestService_Delete_TaskNotFound(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	err := svc.Delete(ctx, 1, 999)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestService_Delete_Forbidden(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()

	task, _ := svc.Create(ctx, 1, "Task")

	err := svc.Delete(ctx, 2, task.ID)
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
