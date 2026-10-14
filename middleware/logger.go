package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Run the remaining middleware and the route handler.
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		status := c.Writer.Status()
		duration := time.Since(start).Seconds()

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		logger.Log(c.Request.Context(), level, "request completed",
			"method", c.Request.Method,
			"route", route,
			"status", status,
			"duration_seconds", duration,
		)
	}
}
