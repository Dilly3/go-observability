
-- +goose Up

CREATE TABLE wallets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_name TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'NGN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT wallets_owner_valid
        CHECK (char_length(btrim(owner_name)) BETWEEN 1 AND 100),

    CONSTRAINT wallets_currency_supported
        CHECK (currency = 'NGN')
);

CREATE TABLE accounts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    account_type TEXT NOT NULL,
    wallet_id BIGINT UNIQUE REFERENCES wallets(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT accounts_type_valid
        CHECK (account_type IN ('asset', 'liability')),

    CONSTRAINT accounts_wallet_relationship
        CHECK (
            (account_type = 'asset' AND wallet_id IS NULL)
            OR
            (account_type = 'liability' AND wallet_id IS NOT NULL)
        )
);

CREATE TABLE journal_transactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    idempotency_key TEXT NOT NULL UNIQUE,
    request_hash TEXT NOT NULL,
    transaction_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT journal_idempotency_key_valid
        CHECK (
            char_length(btrim(idempotency_key)) BETWEEN 1 AND 128
        ),

    CONSTRAINT journal_request_hash_valid
        CHECK (request_hash ~ '^[0-9a-f]{64}$'),

    CONSTRAINT journal_type_valid
        CHECK (transaction_type IN ('funding', 'transfer'))
);

CREATE TABLE ledger_entries (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    journal_transaction_id BIGINT NOT NULL
        REFERENCES journal_transactions(id),
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    direction TEXT NOT NULL,
    amount_kobo BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ledger_direction_valid
        CHECK (direction IN ('debit', 'credit')),

    CONSTRAINT ledger_amount_positive
        CHECK (amount_kobo > 0),

    CONSTRAINT ledger_one_entry_per_account
        UNIQUE (journal_transaction_id, account_id)
);

CREATE INDEX ledger_entries_account_history_idx
    ON ledger_entries (account_id, id);

INSERT INTO accounts (code, account_type)
VALUES ('platform_cash', 'asset');