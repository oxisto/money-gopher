-- name: LinkTransactionDocument :exec
INSERT INTO
    transaction_documents (transaction_id, document_id)
VALUES
    (?, ?) ON CONFLICT DO NOTHING;

-- name: ListTransactionDocuments :many
-- The documents describing a transaction, oldest link first: without the
-- file and the extracted text, which a list of transactions never needs.
SELECT
    documents.id,
    documents.filename,
    documents.content_type,
    documents.state,
    documents.detected_bank,
    documents.error,
    documents.settlement_iban,
    documents.transaction_date,
    documents.created_at,
    documents.person_id,
    documents.content_hash
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

-- name: ListTransactionDocumentsByCashAccount :many
-- The document links of every transaction on an account, in one go.
SELECT
    transaction_documents.transaction_id,
    documents.id,
    documents.filename,
    documents.content_type,
    documents.state,
    documents.detected_bank,
    documents.error,
    documents.settlement_iban,
    documents.transaction_date,
    documents.created_at,
    documents.person_id,
    documents.content_hash
FROM
    transaction_documents
    JOIN transactions ON transactions.id = transaction_documents.transaction_id
    JOIN documents ON documents.id = transaction_documents.document_id
WHERE
    transactions.cash_account_id = ?
ORDER BY
    transaction_documents.linked_at;
