package tasks

import (
	"fmt"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"todo-backend/internal/users"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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

type aiResponse struct {
	Summary  string `json:"summary"`
	Priority string `json:"priority"`
	Tags     string `json:"tags"`
}

func (s *Service) AnalyzeTask(ctx context.Context, userID uint, taskID uint) (*Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, ErrForbidden
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not set in environment")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	prompt := `You are a productivity assistant. Analyze the task and return ONLY a JSON object with three string fields: "summary" (short advice), "priority" (Low, Medium, or High), "tags" (2-3 comma-separated keywords). Task: "` + task.Title + `"`

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("empty response from Gemini")
	}

	part := resp.Candidates[0].Content.Parts[0]
	aiText := fmt.Sprintf("%v", part)

	var result aiResponse
	if err := json.Unmarshal([]byte(aiText), &result); err != nil {
		task.AIGeneratedSummary = "AI Analysis: " + aiText
		task.AIPriority = "Medium"
		task.AITags = "ai-analyzed"
	} else {
		task.AIGeneratedSummary = result.Summary
		task.AIPriority = result.Priority
		task.AITags = result.Tags
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}