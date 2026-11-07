package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper func(c *gin.Context) bool
	Limit   int
	Burst   int
}

var DefaultRateLimitConfig = RateLimitConfig{
	Skipper: DefaultSkipper,
	Limit:   2,
	Burst:   1,
}

func DefaultSkipper(c *gin.Context) bool {
	return false
}

func RateLimitMiddleware() gin.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig)
}

func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	limiter := rate.NewLimiter(rate.Limit(config.Limit), config.Burst)

	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too Many Requests",
			})
			return
		}
		c.Next()
	}
}
