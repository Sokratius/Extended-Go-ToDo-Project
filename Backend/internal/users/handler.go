package users

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"todo-backend/pkg/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/register", h.register)
	router.POST("/login", h.login)
}

type authRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONValidationError(c, "invalid input: username and password are required")
		return
	}

	user, err := h.service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidUsername), errors.Is(err, ErrInvalidPassword):
			utils.JSONValidationError(c, err.Error())
		case errors.Is(err, ErrUsernameTaken):
			utils.JSONError(c, http.StatusConflict, err.Error())
		default:
			utils.JSONError(c, http.StatusInternalServerError, "failed to register user")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

func (h *Handler) login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONValidationError(c, "invalid input: username and password are required")
		return
	}

	user, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			utils.JSONError(c, http.StatusUnauthorized, err.Error())
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "failed to login")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"message":  "login successful; use X-User-ID header for /tasks",
	})
}
