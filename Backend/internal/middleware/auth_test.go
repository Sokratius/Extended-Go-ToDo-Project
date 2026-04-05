package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"todo-backend/internal/middleware"
	"todo-backend/pkg/utils"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthRequired_ValidToken(t *testing.T) {
	token, err := utils.GenerateToken(42, "alice")
	assert.NoError(t, err)

	var capturedUserID uint
	r := gin.New()
	r.Use(middleware.AuthRequired())
	r.GET("/test", func(c *gin.Context) {
		capturedUserID = middleware.GetUserID(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(42), capturedUserID)
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	r := gin.New()
	r.Use(middleware.AuthRequired())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Contains(t, body["error"], "missing")
}

func TestAuthRequired_InvalidFormat(t *testing.T) {
	r := gin.New()
	r.Use(middleware.AuthRequired())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidTokenNoBearer")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_ExpiredOrBadToken(t *testing.T) {
	r := gin.New()
	r.Use(middleware.AuthRequired())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer totally.invalid.token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Contains(t, body["error"], "invalid")
}

func TestGetUserID_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := middleware.GetUserID(c)
	assert.Equal(t, uint(0), userID)
}
