-- name: LinkTransactionDocument :exec
INSERT INTO
    transaction_documents (transaction_id, document_id)
VALUES
    (?, ?) ON CONFLICT DO NOTHING;

-- name: ListTransactionDocuments :many
-- The documents describing a transaction, oldest link first.
SELECT
    documents.*
FROM
    documents
    JOIN transaction_documents ON transaction_documents.document_id = documents.id
WHERE
    transaction_documents.transaction_id = ?
ORDER BY
    transaction_documents.linked_at;

-- name: CountTransactionDocuments :one
SELECT
    count(*)
FROM
    transaction_documents
WHERE
    transaction_id = ?;

-- name: ListUnconfirmedBySchedule :many
-- A schedule's bookings no document describes yet.
SELECT
    transactions.*
FROM
    transactions
    LEFT JOIN transaction_documents ON transaction_documents.transaction_id = transactions.id
WHERE
    transactions.schedule_id = ?
    AND transaction_documents.transaction_id IS NULL
ORDER BY
    transactions.scheduled_for;
