-- +goose Up

-- Link a committed transaction back to the document it was imported from,
-- so a transaction can be traced back to the original bank confirmation
-- (a PDF, a CSV row) that produced it. `source` already records *how* a
-- transaction was created ('MANUAL', 'import:ing', ...) but not *which*
-- specific document — this is the specific link. NULL for manually created
-- transactions and for every transaction that existed before this column.
ALTER TABLE transactions ADD COLUMN document_id TEXT REFERENCES documents(id) ON DELETE SET NULL;

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
