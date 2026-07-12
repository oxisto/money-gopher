-- name: CreateDocument :one
INSERT INTO
    documents (id, user_id, filename, content_type, data)
VALUES
    (?, ?, ?, ?, ?) RETURNING *;

-- name: GetDocument :one
SELECT
    *
FROM
    documents
WHERE
    id = ?
    AND user_id = ?;

-- name: ListDocuments :many
SELECT
    *
FROM
    documents
WHERE
    user_id = ?
ORDER BY
    transaction_date ASC NULLS LAST, created_at ASC;

-- name: ListDocumentsByState :many
SELECT
    *
FROM
    documents
WHERE
    user_id = ?
    AND state = ?
ORDER BY
    transaction_date ASC NULLS LAST, created_at ASC;

-- name: UpdateDocumentState :one
UPDATE documents
SET
    state = ?,
    detected_bank = ?,
    extracted_text = ?,
    error = ?,
    settlement_iban = ?,
    transaction_date = ?
WHERE
    id = ? RETURNING *;

-- name: DeleteDocument :execrows
DELETE FROM documents
WHERE
    id = ?
    AND user_id = ?;

-- name: CreateStagedTransaction :one
INSERT INTO
    staged_transactions (
        id,
        document_id,
        type,
        time,
        units,
        price,
        fees,
        taxes,
        cash_delta,
        currency,
        security_hint,
        isin,
        security_id
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: ListStagedTransactions :many
SELECT
    *
FROM
    staged_transactions
WHERE
    document_id = ?
ORDER BY
    time;

-- name: GetStagedTransaction :one
SELECT
    *
FROM
    staged_transactions
WHERE
    id = ?;

-- name: DeleteStagedTransactions :exec
DELETE FROM staged_transactions
WHERE
    document_id = ?;
