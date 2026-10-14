package datastore

import (
	"context"
	"fmt"
	"time"

	"github.com/dilly3/go-observability/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletStore struct {
	pool *pgxpool.Pool
}

// func WalletStore(pool *pgxpool.Pool) *WalletStore {
// 	return &WalletStore{pool: pool}
// }

func (s *WalletStore) Create(ctx context.Context, ownerName string) (models.Wallet, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("begin wallet transaction: %w", err)
	}

	// Clean up an unfinished transaction, even if the request was canceled.
	// After a successful commit, Rollback has no effect.
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_ = tx.Rollback(cleanupCtx)
	}()

	var wallet models.Wallet
	var currency string

	err = tx.QueryRow(ctx, `
		INSERT INTO wallets (owner_name)
		VALUES ($1)
		RETURNING id::text, owner_name, currency, created_at
	`, ownerName).Scan(
		&wallet.ID,
		&wallet.OwnerName,
		&currency,
		&wallet.CreatedAt,
	)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("insert wallet: %w", err)
	}

	wallet.Currency = models.Currency(currency)

	_, err = tx.Exec(ctx, `
		INSERT INTO accounts (code, account_type, wallet_id)
		VALUES ($1, $2, $3::uuid)
	`,
		"wallet:"+wallet.ID,
		string(models.AccountTypeLiability),
		wallet.ID,
	)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("insert wallet account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Wallet{}, fmt.Errorf("commit wallet transaction: %w", err)
	}

	return wallet, nil
}
