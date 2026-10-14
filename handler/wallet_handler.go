package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) CreateWallet(c *gin.Context) {
	var input struct {
		OwnerName string `json:"owner_name"`
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "expected JSON containing owner_name",
		})
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "expected exactly one JSON object",
		})
		return
	}

	input.OwnerName = strings.TrimSpace(input.OwnerName)
	length := utf8.RuneCountInString(input.OwnerName)

	if length < 1 || length > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "owner_name must contain 1 to 100 characters",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	start := time.Now()
	wallet, err := h.store.WalletStore().Create(ctx, input.OwnerName)

	if err != nil {
		h.logger.Error("wallet creation failed",
			"error", err,
			"duration_seconds", time.Since(start).Seconds(),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create wallet",
		})
		return
	}

	h.logger.Info("wallet created",
		"wallet_id", wallet.ID,
		"duration_seconds", time.Since(start).Seconds(),
	)

	c.Header("Location", "/wallets/"+wallet.ID)
	c.JSON(http.StatusCreated, wallet)
}

// Accept the canonical UUID representation used by our API.
var walletIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

func (h *Handler) GetWallet(c *gin.Context) {
	walletID := c.Param("id")

	if !walletIDPattern.MatchString(walletID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wallet id must be a UUID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	start := time.Now()
	wallet, err := h.store.WalletStore().Get(ctx, walletID)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "wallet not found",
		})
		return
	}

	if err != nil {
		h.logger.Error("wallet retrieval failed",
			"wallet_id", walletID,
			"error", err,
			"duration_seconds", time.Since(start).Seconds(),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to retrieve wallet",
		})
		return
	}

	c.JSON(http.StatusOK, wallet)
}
