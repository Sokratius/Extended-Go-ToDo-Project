package tasks

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todo-backend/internal/users"
	//"todo-backend/internal/middleware"
	"todo-backend/pkg/utils"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/tasks", h.list)
	router.POST("/tasks", h.create)
	router.PUT("/tasks/:id", h.update)
	router.DELETE("/tasks/:id", h.delete)
	router.POST("/tasks/:id/analyze", h.Analyze)
}

type createTaskRequest struct {
	Title string `json:"title" binding:"required"`
}

type updateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

type errorResponse struct {
	Message string `json:"message"`
}

// @Summary Get Tasks
// @Description  Get list of tasks
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Success      200  {object} map[string][]tasks.Task
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /tasks [get]
// @Security ApiKeyAuth
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

// @Summary Create Task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        input body createTaskRequest true "create task"
// @Success      201  {object} map[string][]tasks.Task "new task"
// @Failure      401  {object} errorResponse
// @Failure      500  {object} errorResponse
// @Router       /tasks [post]
// @Security ApiKeyAuth
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

// @Summary Update Task
// @Description  Update Task Fields
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id    path int true "Task ID"
// @Param        input body updateTaskRequest true "Update Task"
// @Success      200  {object} map[string][]tasks.Task "updated_task"
// @Failure      401  {object} errorResponse
// @Failure      403  {object} errorResponse
// @Failure      404  {object} errorResponse
// @Failure      500  {object} errorResponse
// @Router       /tasks/{id} [put]
// @Security ApiKeyAuth
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

// @Summary Delete Task
// @Description  Delete task by ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param 		 id path int true "Task ID"
// @Success      200  {object} map[string]string
// @Failure      401  {object} errorResponse
// @Failure      500  {object} errorResponse
// @Router       /tasks/{id} [delete]
// @Security ApiKeyAuth
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

// @Summary Analyze Task with AI
// @Description Send task to AI to generate summary, tags, and priority
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} tasks.Task "analyzed_task"
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /tasks/{id}/analyze [post]
// @Security BearerAuth
func (h *Handler) Analyze(c *gin.Context) {
	userID := uint(1)

	taskID, ok := parseTaskID(c)
	if !ok {
		return
	}

	task, err := h.service.AnalyzeTask(c.Request.Context(), userID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTaskNotFound):
			utils.JSONError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			utils.JSONError(c, http.StatusForbidden, err.Error())
		default:
			utils.JSONError(c, http.StatusInternalServerError, "failed to analyze task")
		}
		return
	}

	c.JSON(http.StatusOK, task)
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

func (h *Handler) userIDFromHeader(c *gin.Context) (uint, bool) {
	userID, err := utils.ParseUserIDHeader(c)
	if err != nil {
		utils.JSONValidationError(c, err.Error())
		return 0, false
	}
	return userID, true
}

