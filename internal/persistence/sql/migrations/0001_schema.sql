-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    -- User is a person using money-gopher, provisioned on first OIDC login.
    id           TEXT     PRIMARY KEY,
    issuer       TEXT     NOT NULL,
    subject      TEXT     NOT NULL,
    display_name TEXT     NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT (now()),
    UNIQUE (issuer, subject)
);

CREATE TABLE IF NOT EXISTS cash_accounts (
    -- CashAccount is a real-world bank/cash account. Its balance is derived
    -- as the sum of all transaction cash_deltas settling against it.
    id           TEXT     PRIMARY KEY,
    -- Superseded by person_id, added in migration 0002. Kept only because
    -- lightsql cannot DROP COLUMN; never populated by new rows.
    user_id      TEXT     REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT     NOT NULL,
    currency     TEXT     NOT NULL,
    iban         TEXT,
    created_at   TIMESTAMP NOT NULL DEFAULT (now())
);

CREATE TABLE IF NOT EXISTS portfolios (
    -- Portfolio is a collection of security positions belonging to one user.
    -- Every portfolio settles trades and dividends against exactly one cash
    -- account; the relationship is mandatory.
    id                TEXT     PRIMARY KEY,
    -- Superseded by person_id, added in migration 0002. Kept only because
    -- lightsql cannot DROP COLUMN; never populated by new rows.
    user_id           TEXT     REFERENCES users(id) ON DELETE CASCADE,
    display_name      TEXT     NOT NULL,
    cash_account_id   TEXT     NOT NULL REFERENCES cash_accounts(id) ON DELETE RESTRICT,
    created_at        TIMESTAMP NOT NULL DEFAULT (now())
);

CREATE TABLE IF NOT EXISTS securities (
    -- Security is a tradable security, shared across all users.
    id           TEXT PRIMARY KEY,
    display_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS security_identifiers (
    -- SecurityIdentifier maps an external identifier (ISIN, WKN, …) to a security.
    security_id TEXT NOT NULL REFERENCES securities(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL,
    value       TEXT NOT NULL,
    PRIMARY KEY (security_id, kind, value)
);

-- ISINs and WKNs are globally unique; tickers may collide across exchanges.
CREATE UNIQUE INDEX IF NOT EXISTS idx_security_identifiers_unique
    ON security_identifiers(kind, value)
    WHERE kind IN ('ISIN', 'WKN');

CREATE TABLE IF NOT EXISTS listings (
    -- Listing is a security listed on a particular exchange.
    id             TEXT PRIMARY KEY,
    security_id    TEXT NOT NULL REFERENCES securities(id) ON DELETE CASCADE,
    exchange       TEXT,
    ticker         TEXT NOT NULL,
    currency       TEXT NOT NULL,
    quote_provider TEXT,
    UNIQUE (security_id, ticker)
);

CREATE TABLE IF NOT EXISTS quotes (
    -- Quote is one price observation for a listing.
    listing_id TEXT    NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    time       TIMESTAMP NOT NULL,
    price      INTEGER  NOT NULL,
    PRIMARY KEY (listing_id, time)
);

CREATE TABLE IF NOT EXISTS transactions (
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
    price           INTEGER,
    fees            INTEGER  NOT NULL DEFAULT 0,
    taxes           INTEGER  NOT NULL DEFAULT 0,
    cash_delta      INTEGER  NOT NULL DEFAULT 0,
    currency        TEXT     NOT NULL,
    source          TEXT     NOT NULL DEFAULT 'MANUAL',
    created_at      TIMESTAMP NOT NULL DEFAULT (now()),
    CHECK (portfolio_id IS NOT NULL OR cash_account_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_time
    ON transactions(portfolio_id, time);

CREATE INDEX IF NOT EXISTS idx_transactions_cash_account_time
    ON transactions(cash_account_id, time);

CREATE TABLE IF NOT EXISTS documents (
    -- Document is an uploaded file going through the import pipeline.
    id              TEXT     PRIMARY KEY,
    -- Superseded by person_id, added in migration 0002. Kept only because
    -- lightsql cannot DROP COLUMN; never populated by new rows.
    user_id         TEXT     REFERENCES users(id) ON DELETE CASCADE,
    filename        TEXT     NOT NULL,
    content_type    TEXT     NOT NULL,
    data            BYTEA    NOT NULL,
    state           TEXT     NOT NULL DEFAULT 'UPLOADED' CHECK (
                        state IN ('UPLOADED','PARSED','FAILED','SKIPPED','IMPORTED')
                    ),
    detected_bank   TEXT,
    extracted_text  TEXT,
    error           TEXT,
    settlement_iban TEXT,
    transaction_date TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT (now())
);

CREATE TABLE IF NOT EXISTS staged_transactions (
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
    price         INTEGER,
    fees          INTEGER  NOT NULL DEFAULT 0,
    taxes         INTEGER  NOT NULL DEFAULT 0,
    cash_delta    INTEGER  NOT NULL DEFAULT 0,
    currency      TEXT     NOT NULL,
    security_hint TEXT,
    isin          TEXT,
    security_id   TEXT     REFERENCES securities(id) ON DELETE SET NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT (now())
);

CREATE INDEX IF NOT EXISTS idx_staged_transactions_document
    ON staged_transactions(document_id);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT     PRIMARY KEY,
    user_id    TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT (now()),
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_user_id ON sessions(user_id);

-- +goose Down
DROP INDEX IF EXISTS sessions_user_id;
DROP TABLE IF EXISTS sessions;
DROP INDEX IF EXISTS idx_staged_transactions_document;
DROP TABLE IF EXISTS staged_transactions;
DROP TABLE IF EXISTS documents;
DROP INDEX IF EXISTS idx_transactions_cash_account_time;
DROP INDEX IF EXISTS idx_transactions_portfolio_time;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS quotes;
DROP TABLE IF EXISTS listings;
DROP INDEX IF EXISTS idx_security_identifiers_unique;
DROP TABLE IF EXISTS security_identifiers;
DROP TABLE IF EXISTS securities;
DROP TABLE IF EXISTS portfolios;
DROP TABLE IF EXISTS cash_accounts;
DROP TABLE IF EXISTS users;
