-- +goose Up
-- A free text on a transaction: what the statement printed for the row, or
-- what the person wrote. Lists show it as the title when set.
ALTER TABLE transactions ADD COLUMN note TEXT;
ALTER TABLE staged_transactions ADD COLUMN note TEXT;

-- +goose Down
