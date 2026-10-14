package models

import "time"

type Currency string
type AccountType string
type TransactionType string
type EntryDirection string

const (
	CurrencyNGN Currency = "NGN"

	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"

	TransactionTypeFunding  TransactionType = "funding"
	TransactionTypeTransfer TransactionType = "transfer"

	EntryDirectionDebit  EntryDirection = "debit"
	EntryDirectionCredit EntryDirection = "credit"
)

// Wallet identifies an owner. Its balance comes from ledger entries.
type Wallet struct {
	ID        string    `json:"id"`
	OwnerName string    `json:"owner_name"`
	Currency  Currency  `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type Account struct {
	ID          string      `json:"id"`
	Code        string      `json:"code"`
	AccountType AccountType `json:"account_type"`
	WalletID    *string     `json:"wallet_id"`
	CreatedAt   time.Time   `json:"created_at"`
}

// JournalTransaction groups the entries for one money movement.
type JournalTransaction struct {
	ID              string          `json:"id"`
	IdempotencyKey  string          `json:"-"`
	RequestHash     string          `json:"-"`
	TransactionType TransactionType `json:"transaction_type"`
	CreatedAt       time.Time       `json:"created_at"`
}

// LedgerEntry records one side of a journal transaction.
// AmountKobo is always positive; Direction determines its accounting effect.
type LedgerEntry struct {
	ID                   string         `json:"id"`
	JournalTransactionID string         `json:"journal_transaction_id"`
	AccountID            string         `json:"account_id"`
	Direction            EntryDirection `json:"direction"`
	AmountKobo           int64          `json:"amount_kobo"`
	CreatedAt            time.Time      `json:"created_at"`
}
