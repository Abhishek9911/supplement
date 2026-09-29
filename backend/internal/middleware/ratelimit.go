package middleware

import (
	"net/http"
	"sync"
	"time"

	"suppliment/backend/internal/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var limiterStore sync.Map

func RateLimiter(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if value, ok := limiterStore.Load(key); ok {
			l := value.(*rate.Limiter)
			if !l.Allow() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
				return
			}
		} else {
			l := rate.NewLimiter(rate.Every(cfg.RateLimitWindow/time.Duration(cfg.RateLimitMax)), cfg.RateLimitMax)
			limiterStore.Store(key, l)
			if !l.Allow() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
				return
			}
		}
		c.Next()
	}
}
