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

type errorResponse struct {
	Message string `json:"message"`
}

// @Summary Register
// @Tags auth
// @ID create-account
// @Accept json
// @Produce json
// @Param input body authRequest true "account info"
// @Success 200 {integer} integer 1
// @Success 201 {integer} integer 1
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /register [post]
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

// @Summary      Login
// @Description  Authorization with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  authRequest  true "account info"
// @Success      200  {object}  errorResponse
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /login [post]
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
