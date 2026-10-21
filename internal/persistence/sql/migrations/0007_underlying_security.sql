-- +goose Up

-- Links a security to the one it derives from — a subscription right or an
-- option/warrant to its underlying stock. Self-referential and nullable:
-- most securities (an ordinary stock, a bond, an ETF) have no underlying at
-- all. General on purpose rather than rights-specific, so the same column
-- covers calls/puts/warrants later without another migration.
--
-- Plain ADD COLUMN, not a rebuild: unlike the DECIMAL precision migrations,
-- this is a genuinely new nullable column with no existing data to
-- reconcile, which is exactly what lightsql's ALTER TABLE supports natively
-- (same pattern as transactions.document_id in 0004).
ALTER TABLE securities ADD COLUMN underlying_security_id TEXT REFERENCES securities(id) ON DELETE SET NULL;

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
