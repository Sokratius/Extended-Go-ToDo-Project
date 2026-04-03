package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"todo-backend/pkg/utils"
)

const (
	ContextUserID   = "userID"
	ContextUsername = "username"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing Authorization header",
			})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid Authorization header format, expected: Bearer <token>",
			})
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Next()
	}
}

// GetUserID вытаскивает аутентифицированный user ID с gin контекста.
// Возвращает 0 если не установлен.
func GetUserID(c *gin.Context) uint {
	id, _ := c.Get(ContextUserID)
	userID, _ := id.(uint)
	return userID
}
