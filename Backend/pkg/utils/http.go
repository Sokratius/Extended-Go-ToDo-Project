package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseUserIDHeader(c *gin.Context) (uint, error) {
	raw := c.GetHeader("X-User-ID")
	if raw == "" {
		return 0, fmt.Errorf("missing X-User-ID header")
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid X-User-ID header")
	}
	return uint(parsed), nil
}

func JSONError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func JSONValidationError(c *gin.Context, message string) {
	JSONError(c, http.StatusBadRequest, message)
}
