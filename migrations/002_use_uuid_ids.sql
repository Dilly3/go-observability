-- +goose Up

-- Prevent writes while we check and change the schema.
LOCK TABLE wallets, accounts, journal_transactions, ledger_entries
    IN ACCESS EXCLUSIVE MODE;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM wallets)
        OR EXISTS (SELECT 1 FROM journal_transactions)
        OR EXISTS (SELECT 1 FROM ledger_entries)
        OR (SELECT COUNT(*) FROM accounts) <> 1
        OR NOT EXISTS (
            SELECT 1
            FROM accounts
            WHERE code = 'platform_cash'
              AND account_type = 'asset'
              AND wallet_id IS NULL
        )
    THEN
        RAISE EXCEPTION
            'UUID migration requires empty wallets and ledger, with only the platform_cash account';
    END IF;
END;
$$;
-- +goose StatementEnd

-- Temporarily remove references to the integer IDs.
ALTER TABLE accounts
    DROP CONSTRAINT accounts_wallet_id_fkey;

ALTER TABLE ledger_entries
    DROP CONSTRAINT ledger_entries_journal_transaction_id_fkey,
    DROP CONSTRAINT ledger_entries_account_id_fkey;

-- Remove integer identity generation.
ALTER TABLE wallets
    ALTER COLUMN id DROP IDENTITY;

ALTER TABLE accounts
    ALTER COLUMN id DROP IDENTITY;

ALTER TABLE journal_transactions
    ALTER COLUMN id DROP IDENTITY;

ALTER TABLE ledger_entries
    ALTER COLUMN id DROP IDENTITY;

-- Change primary keys to UUIDs.
-- The existing platform account receives a UUID and is preserved.
ALTER TABLE wallets
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE accounts
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE journal_transactions
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE ledger_entries
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- These references are empty or NULL, as verified above.
ALTER TABLE accounts
    ALTER COLUMN wallet_id TYPE UUID USING NULL::UUID;

ALTER TABLE ledger_entries
    ALTER COLUMN journal_transaction_id TYPE UUID USING NULL::UUID,
    ALTER COLUMN account_id TYPE UUID USING NULL::UUID;

-- Restore foreign keys using the new UUID types.
ALTER TABLE accounts
    ADD CONSTRAINT accounts_wallet_id_fkey
    FOREIGN KEY (wallet_id) REFERENCES wallets(id);

ALTER TABLE ledger_entries
    ADD CONSTRAINT ledger_entries_journal_transaction_id_fkey
    FOREIGN KEY (journal_transaction_id)
    REFERENCES journal_transactions(id),

    ADD CONSTRAINT ledger_entries_account_id_fkey
    FOREIGN KEY (account_id) REFERENCES accounts(id);