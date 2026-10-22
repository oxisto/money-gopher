-- +goose Up

-- The SHA-256 of a document's bytes, so uploading the same statement twice
-- is caught at upload rather than ending in duplicate transactions. NULL
-- for documents stored before this existed; moneyd fills them in at
-- startup.
ALTER TABLE documents ADD COLUMN content_hash TEXT;

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
