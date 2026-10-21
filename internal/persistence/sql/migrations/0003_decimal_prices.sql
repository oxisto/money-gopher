-- +goose Up

-- Price columns were changed from INTEGER (minor units) to DECIMAL(19,4)
-- (major units) directly in 0001_schema.sql rather than as a proper
-- migration — fine for a schema that had never shipped, except any database
-- that had already applied 0001 before the edit is stuck: goose tracks
-- migrations by version number, not content, so it never reruns 0001, and
-- lightsql's ALTER TABLE does not support a column type change ("neither can
-- be served by a missing value", per its own feature notes). The fix has to
-- rebuild the three affected tables.
--
-- quotes and transactions are recreated empty (no INSERT ... SELECT): every
-- environment that hit this had zero rows in both — quotes because nothing
-- had synced yet, transactions because they'd already been cleared ahead of
-- reimporting under the new precision. staged_transactions is also recreated
-- empty rather than converted: it only holds unreviewed import drafts, and
-- those drafts were parsed by the old, less precise (2-decimal) code this
-- whole migration exists to fix — reusing them would just carry the
-- precision loss forward. Losing them isn't data loss: cmd/reprocess
-- regenerates them from each document's stored extracted_text.

DROP TABLE quotes;

CREATE TABLE quotes (
    -- Quote is one price observation for a listing.
    listing_id TEXT    NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    time       TIMESTAMP NOT NULL,
    price      DECIMAL(19,4) NOT NULL,
    PRIMARY KEY (listing_id, time)
);

DROP TABLE transactions;

CREATE TABLE transactions (
    -- Transaction is a typed portfolio/cash event.
    id              TEXT     PRIMARY KEY,
    type            TEXT     NOT NULL CHECK (type IN (
                        'BUY','SELL','DELIVERY_INBOUND','DELIVERY_OUTBOUND',
                        'DIVIDEND','INTEREST','DEPOSIT_CASH','WITHDRAW_CASH',
                        'ACCOUNT_FEES','TAX_REFUND'
                    )),
    time            TIMESTAMP NOT NULL,
    portfolio_id    TEXT     REFERENCES portfolios(id)    ON DELETE CASCADE,
    security_id     TEXT     REFERENCES securities(id)    ON DELETE RESTRICT,
    cash_account_id TEXT     REFERENCES cash_accounts(id) ON DELETE CASCADE,
    units           REAL     NOT NULL DEFAULT 0,
    price           DECIMAL(19,4),
    fees            INTEGER  NOT NULL DEFAULT 0,
    taxes           INTEGER  NOT NULL DEFAULT 0,
    cash_delta      INTEGER  NOT NULL DEFAULT 0,
    currency        TEXT     NOT NULL,
    source          TEXT     NOT NULL DEFAULT 'MANUAL',
    created_at      TIMESTAMP NOT NULL DEFAULT (now()),
    CHECK (portfolio_id IS NOT NULL OR cash_account_id IS NOT NULL)
);

CREATE INDEX idx_transactions_portfolio_time
    ON transactions(portfolio_id, time);

CREATE INDEX idx_transactions_cash_account_time
    ON transactions(cash_account_id, time);

DROP TABLE staged_transactions;

CREATE TABLE staged_transactions (
    -- StagedTransaction is a parsed transaction awaiting user review.
    id            TEXT     PRIMARY KEY,
    document_id   TEXT     NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    type          TEXT     NOT NULL CHECK (type IN (
                      'BUY','SELL','DELIVERY_INBOUND','DELIVERY_OUTBOUND',
                      'DIVIDEND','INTEREST','DEPOSIT_CASH','WITHDRAW_CASH',
                      'ACCOUNT_FEES','TAX_REFUND'
                  )),
    time          TIMESTAMP NOT NULL,
    units         REAL     NOT NULL DEFAULT 0,
    price         DECIMAL(19,4),
    fees          INTEGER  NOT NULL DEFAULT 0,
    taxes         INTEGER  NOT NULL DEFAULT 0,
    cash_delta    INTEGER  NOT NULL DEFAULT 0,
    currency      TEXT     NOT NULL,
    security_hint TEXT,
    isin          TEXT,
    security_id   TEXT     REFERENCES securities(id) ON DELETE SET NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT (now())
);

CREATE INDEX idx_staged_transactions_document
    ON staged_transactions(document_id);

-- Every document whose staged transactions were just discarded needs to go
-- back through the import pipeline; UPLOADED is where a fresh document
-- starts. IMPORTED documents are untouched (nothing here dropped their
-- committed transactions), and there were none in this state anywhere this
-- bug was hit.
UPDATE documents SET state = 'UPLOADED' WHERE state = 'PARSED';

-- +goose Down

-- Not reversible: the old INTEGER-minor-units price data was already
-- dropped by the Up migration (see above), and lightsql cannot add back a
-- narrower column type either.
