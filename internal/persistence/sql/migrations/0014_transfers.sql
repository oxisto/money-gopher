-- +goose Up

-- A transfer between two own accounts is one event with two legs: a
-- WITHDRAW_CASH on the account it leaves and a DEPOSIT_CASH on the one it
-- reaches, sharing a transfer_id. No new transaction type: the type list
-- is a CHECK, and lightsql cannot rebuild the table to change it.
ALTER TABLE transactions ADD COLUMN transfer_id TEXT;
CREATE INDEX idx_transactions_transfer ON transactions(transfer_id);

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
