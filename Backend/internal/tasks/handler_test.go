package tasks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"todo-backend/internal/tasks"
	"todo-backend/internal/users"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) ListByUser(ctx context.Context, userID uint) ([]tasks.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]tasks.Task), args.Error(1)
}

func (m *mockService) Create(ctx context.Context, userID uint, title string) (*tasks.Task, error) {
	args := m.Called(ctx, userID, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tasks.Task), args.Error(1)
}

func (m *mockService) Update(ctx context.Context, userID, taskID uint, title *string, done *bool) (*tasks.Task, error) {
	args := m.Called(ctx, userID, taskID, title, done)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tasks.Task), args.Error(1)
}

func (m *mockService) Delete(ctx context.Context, userID, taskID uint) error {
	args := m.Called(ctx, userID, taskID)
	return args.Error(0)
}

// setupRouter создаёт gin роутер в тестовом режиме.
// Handler принимает интерфейс — если у тебя сейчас *Service, нужно будет
// вынести интерфейс (см. комментарий внизу файла).
func setupRouter(svc tasks.ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := tasks.NewHandler(svc)
	h.RegisterRoutes(r)
	return r
}

func ptr[T any](v T) *T { return &v }

func TestList_Success(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	expected := []tasks.Task{
		{ID: 1, UserID: 42, Title: "Buy milk", Done: false},
		{ID: 2, UserID: 42, Title: "Go gym", Done: true},
	}
	svc.On("ListByUser", mock.Anything, uint(42)).Return(expected, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string][]tasks.Task
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Len(t, body["tasks"], 2)
	assert.Equal(t, "Buy milk", body["tasks"][0].Title)
	svc.AssertExpectations(t)
}

func TestList_MissingHeader(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	// заголовок X-User-ID не передаём
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "ListByUser")
}

func TestList_UserNotFound(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("ListByUser", mock.Anything, uint(99)).Return([]tasks.Task{}, users.ErrUserNotFound)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("X-User-ID", "99")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	svc.AssertExpectations(t)
}

func TestList_InternalError(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("ListByUser", mock.Anything, uint(1)).Return([]tasks.Task{}, assert.AnError)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("X-User-ID", "1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestCreate_Success(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	created := &tasks.Task{ID: 10, UserID: 42, Title: "Write tests", Done: false}
	svc.On("Create", mock.Anything, uint(42), "Write tests").Return(created, nil)

	body := `{"title":"Write tests"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var got tasks.Task
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, "Write tests", got.Title)
	svc.AssertExpectations(t)
}

func TestCreate_MissingTitle(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestCreate_InvalidTitle(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("Create", mock.Anything, uint(42), "").Return(nil, tasks.ErrInvalidTitle)

	// Тут binding:"required" уже отсечёт пустой title раньше сервиса,
	// поэтому тест проверяет что статус 400 в любом случае.
	body := `{"title":""}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreate_UserNotFound(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("Create", mock.Anything, uint(42), "Task").Return(nil, users.ErrUserNotFound)

	body := `{"title":"Task"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	svc.AssertExpectations(t)
}

func TestUpdate_Success(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	updated := &tasks.Task{ID: 5, UserID: 42, Title: "Updated", Done: true}
	svc.On("Update", mock.Anything, uint(42), uint(5), ptr("Updated"), ptr(true)).Return(updated, nil)

	body := `{"title":"Updated","done":true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/5", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got tasks.Task
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", got.Title)
	assert.True(t, got.Done)
	svc.AssertExpectations(t)
}

func TestUpdate_NoFields(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/5", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Update")
}

func TestUpdate_InvalidTaskID(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	body := `{"title":"x"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Update")
}

func TestUpdate_TaskNotFound(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("Update", mock.Anything, uint(42), uint(999), ptr("x"), (*bool)(nil)).Return(nil, tasks.ErrTaskNotFound)

	body := `{"title":"x"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestUpdate_Forbidden(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("Update", mock.Anything, uint(42), uint(7), ptr("x"), (*bool)(nil)).Return(nil, tasks.ErrForbidden)

	body := `{"title":"x"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/7", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertExpectations(t)
}

func TestDelete_Success(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("Delete", mock.Anything, uint(42), uint(3)).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/3", nil)
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "deleted", body["status"])
	svc.AssertExpectations(t)
}

func TestDelete_TaskNotFound(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("Delete", mock.Anything, uint(42), uint(999)).Return(tasks.ErrTaskNotFound)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/999", nil)
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestDelete_Forbidden(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	svc.On("Delete", mock.Anything, uint(42), uint(7)).Return(tasks.ErrForbidden)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/7", nil)
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertExpectations(t)
}

func TestDelete_InvalidTaskID(t *testing.T) {
	svc := new(mockService)
	router := setupRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/abc", nil)
	req.Header.Set("X-User-ID", "42")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Delete")
}
