package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dilly3/go-observability/datastore"
	"github.com/dilly3/go-observability/handler"
	"github.com/dilly3/go-observability/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	pool, err := datastore.OpenDb(startupCtx, os.Getenv("DATABASE_URL"))
	cancel()

	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}

	defer pool.Close()

	logger.Info("database connected", "max_connections", pool.Config().MaxConns)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(logger))

	router.GET("/healthz", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})

	})

	router.GET("/readyz", handler.ReadinessHandler(pool, logger))

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("server starting", "address", server.Addr)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
