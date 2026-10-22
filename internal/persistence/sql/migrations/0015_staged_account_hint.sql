-- +goose Up
-- Which of a document's accounts a staged row belongs to, for documents
-- that carry more than one statement (an LBS year-end bundle lists the
-- Bausparvertrag and the loan side by side).
ALTER TABLE staged_transactions ADD COLUMN account_hint TEXT;

-- +goose Down
