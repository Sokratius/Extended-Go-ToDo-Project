package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"todo-backend/pkg/utils" 
)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 3)
		visitors[ip] = limiter
	}
	return limiter
}

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip)

		if !limiter.Allow() {
			utils.JSONError(c, http.StatusTooManyRequests, "Слишком много запросов. Пожалуйста, подождите.")
			c.Abort()
			return
		}
		c.Next()
	}
}