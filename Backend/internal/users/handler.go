package users

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"todo-backend/internal/middleware"

	"todo-backend/pkg/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/register", middleware.RateLimiter(), h.register)
	router.POST("/login", middleware.RateLimiter(), h.login)
}

type authRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	Token    string `json:"token"`
	ID       uint   `json:"id"`
	Username string `json:"username"`
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
// @Success 201 {object} authResponse
// @Failure 400 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
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

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		Token:    token,
		ID:       user.ID,
		Username: user.Username,
	})
}

// @Summary      Login
// @Description  Authorization with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  authRequest  true "account info"
// @Success      200  {object}  authResponse
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

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	c.JSON(http.StatusOK, authResponse{
		Token:    token,
		ID:       user.ID,
		Username: user.Username,
	})
}
