package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
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
	wallet, err := h.store.GetWalletStore().Create(ctx, input.OwnerName)

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
