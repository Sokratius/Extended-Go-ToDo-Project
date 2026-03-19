package tasks

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todo-backend/internal/users"
	"todo-backend/pkg/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/tasks", h.list)
	router.POST("/tasks", h.create)
	router.PUT("/tasks/:id", h.update)
	router.DELETE("/tasks/:id", h.delete)
}

type createTaskRequest struct {
	Title string `json:"title" binding:"required"`
}

type updateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func (h *Handler) list(c *gin.Context) {
	userID, ok := h.userIDFromHeader(c)
	if !ok {
		return
	}

	items, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			utils.JSONError(c, http.StatusUnauthorized, "user not found")
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": items})
}

func (h *Handler) create(c *gin.Context) {
	userID, ok := h.userIDFromHeader(c)
	if !ok {
		return
	}

	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONValidationError(c, "invalid input: title is required")
		return
	}

	task, err := h.service.Create(c.Request.Context(), userID, req.Title)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTitle):
			utils.JSONValidationError(c, err.Error())
		case errors.Is(err, users.ErrUserNotFound):
			utils.JSONError(c, http.StatusUnauthorized, "user not found")
		default:
			utils.JSONError(c, http.StatusInternalServerError, "failed to create task")
		}
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *Handler) update(c *gin.Context) {
	userID, ok := h.userIDFromHeader(c)
	if !ok {
		return
	}

	taskID, ok := parseTaskID(c)
	if !ok {
		return
	}

	var req updateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONValidationError(c, "invalid input")
		return
	}

	if req.Title == nil && req.Done == nil {
		utils.JSONValidationError(c, "at least one field (title or done) must be provided")
		return
	}

	task, err := h.service.Update(c.Request.Context(), userID, taskID, req.Title, req.Done)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTitle):
			utils.JSONValidationError(c, err.Error())
		case errors.Is(err, ErrTaskNotFound):
			utils.JSONError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			utils.JSONError(c, http.StatusForbidden, err.Error())
		default:
			utils.JSONError(c, http.StatusInternalServerError, "failed to update task")
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) delete(c *gin.Context) {
	userID, ok := h.userIDFromHeader(c)
	if !ok {
		return
	}

	taskID, ok := parseTaskID(c)
	if !ok {
		return
	}

	err := h.service.Delete(c.Request.Context(), userID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTaskNotFound):
			utils.JSONError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			utils.JSONError(c, http.StatusForbidden, err.Error())
		default:
			utils.JSONError(c, http.StatusInternalServerError, "failed to delete task")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *Handler) userIDFromHeader(c *gin.Context) (uint, bool) {
	userID, err := utils.ParseUserIDHeader(c)
	if err != nil {
		utils.JSONValidationError(c, err.Error())
		return 0, false
	}
	return userID, true
}

func parseTaskID(c *gin.Context) (uint, bool) {
	rawID := c.Param("id")
	parsed, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || parsed == 0 {
		utils.JSONValidationError(c, "invalid task id")
		return 0, false
	}
	return uint(parsed), true
}
