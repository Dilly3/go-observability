package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/dilly3/go-observability/datastore"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	store  *datastore.Store
	logger *slog.Logger
}

func NewHandler(store *datastore.Store, logger *slog.Logger) Handler {
	return Handler{store: store, logger: logger}
}

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
