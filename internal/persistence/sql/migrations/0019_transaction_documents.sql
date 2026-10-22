-- +goose Up
-- Every document that describes a transaction: the one it was imported
-- from, and every statement that confirmed it since. transactions.document_id
-- stays as the origin.
CREATE TABLE transaction_documents (
    transaction_id TEXT      NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    document_id    TEXT      NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    linked_at      TIMESTAMP NOT NULL DEFAULT (now()),
    PRIMARY KEY (transaction_id, document_id)
);

INSERT INTO transaction_documents (transaction_id, document_id)
SELECT id, document_id FROM transactions WHERE document_id IS NOT NULL;

-- +goose Down
