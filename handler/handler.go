package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ReadinessHandler(pool *pgxpool.Pool, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()

		start := time.Now()
		err := pool.Ping(ctx)

		if err != nil {
			logger.Error("database readiness check failed",
				"error", err,
				"duration_seconds", time.Since(start).Seconds(),
			)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}
