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

func (s *WalletStore) Get(ctx context.Context, walletID string) (models.WalletDetails, error) {
	var result models.WalletDetails
	var currency string
	var accountID *string

	err := s.pool.QueryRow(ctx, `
		SELECT
			w.id::text,
			w.owner_name,
			w.currency,
			w.created_at,
			a.id::text,
			COALESCE(
				SUM(
					CASE
						WHEN e.direction = 'credit' THEN e.amount_kobo
						WHEN e.direction = 'debit' THEN -e.amount_kobo
						ELSE 0
					END
				),
				0
			)::bigint AS balance_kobo
		FROM wallets AS w
		LEFT JOIN accounts AS a
			ON a.wallet_id = w.id
			AND a.account_type = 'liability'
		LEFT JOIN ledger_entries AS e
			ON e.account_id = a.id
		WHERE w.id = $1::uuid
		GROUP BY w.id, a.id
	`, walletID).Scan(
		&result.ID,
		&result.OwnerName,
		&currency,
		&result.CreatedAt,
		&accountID,
		&result.BalanceKobo,
	)
	if err != nil {
		return models.WalletDetails{}, fmt.Errorf("get wallet: %w", err)
	}

	if accountID == nil {
		return models.WalletDetails{}, fmt.Errorf(
			"wallet %s has no liability account",
			walletID,
		)
	}

	result.Currency = models.Currency(currency)

	return result, nil
}
