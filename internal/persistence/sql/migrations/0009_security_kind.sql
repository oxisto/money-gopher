-- +goose Up

-- What a security is: a stock, an ETF, a fund, a bond, or something else.
-- NULL for the ones created before this existed, until someone sets it.
-- The import pipeline suggests one from the document (a bond trades at a
-- percentage of a face value; "ETF" in the name says enough), stored on
-- the staged transaction so the review dialog can preselect it.
ALTER TABLE securities ADD COLUMN kind TEXT CHECK (kind IN ('STOCK', 'ETF', 'FUND', 'BOND', 'OTHER'));
ALTER TABLE staged_transactions ADD COLUMN security_kind TEXT;

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
