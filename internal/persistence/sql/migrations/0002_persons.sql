-- +goose NO TRANSACTION
-- +goose Up

-- This migration must run outside a transaction: PRAGMA foreign_keys is a
-- no-op inside one, and the table recreation below depends on it being OFF.

CREATE TABLE persons (
    id           TEXT     PRIMARY KEY,
    display_name TEXT     NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))
);

-- user_person_access grants a user the ability to view and manage a person's
-- financial data. One user can access many persons; a person can be accessed
-- by many users (e.g. a shared household view).
CREATE TABLE user_person_access (
    user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    person_id TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, person_id)
);

-- Create a person for every existing user (same id for a seamless migration).
INSERT INTO persons (id, display_name, created_at)
SELECT id, display_name, created_at FROM users;

INSERT INTO user_person_access (user_id, person_id)
SELECT id, id FROM users;

-- Add active person to sessions; NULL means "default to first accessible person".
ALTER TABLE sessions ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE SET NULL;

-- Recreate portfolios with person_id instead of user_id.
PRAGMA foreign_keys = OFF;

CREATE TABLE portfolios_new (
    id              TEXT     PRIMARY KEY,
    person_id       TEXT     NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    display_name    TEXT     NOT NULL,
    cash_account_id TEXT     NOT NULL REFERENCES cash_accounts(id) ON DELETE RESTRICT,
    created_at      DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))
);
INSERT INTO portfolios_new SELECT id, user_id, display_name, cash_account_id, created_at FROM portfolios;
DROP TABLE portfolios;
ALTER TABLE portfolios_new RENAME TO portfolios;

CREATE TABLE cash_accounts_new (
    id           TEXT     PRIMARY KEY,
    person_id    TEXT     NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    display_name TEXT     NOT NULL,
    currency     TEXT     NOT NULL,
    iban         TEXT,
    created_at   DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))
);
INSERT INTO cash_accounts_new SELECT id, user_id, display_name, currency, iban, created_at FROM cash_accounts;
DROP TABLE cash_accounts;
ALTER TABLE cash_accounts_new RENAME TO cash_accounts;

CREATE TABLE documents_new (
    id               TEXT     PRIMARY KEY,
    person_id        TEXT     NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    filename         TEXT     NOT NULL,
    content_type     TEXT     NOT NULL,
    data             BLOB     NOT NULL,
    state            TEXT     NOT NULL DEFAULT 'UPLOADED' CHECK (
                         state IN ('UPLOADED','PARSED','FAILED','SKIPPED','IMPORTED')
                     ),
    detected_bank    TEXT,
    extracted_text   TEXT,
    error            TEXT,
    settlement_iban  TEXT,
    transaction_date DATETIME,
    created_at       DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))
);
INSERT INTO documents_new SELECT id, user_id, filename, content_type, data, state,
    detected_bank, extracted_text, error, settlement_iban, transaction_date, created_at
FROM documents;
DROP TABLE documents;
ALTER TABLE documents_new RENAME TO documents;

PRAGMA foreign_keys = ON;

-- Restore indexes dropped by table recreation (use IF NOT EXISTS since the
-- indexes on transactions/staged_transactions/sessions survive the recreation).
CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_time  ON transactions(portfolio_id, time);
CREATE INDEX IF NOT EXISTS idx_transactions_cash_account_time ON transactions(cash_account_id, time);
CREATE INDEX IF NOT EXISTS idx_staged_transactions_document  ON staged_transactions(document_id);
CREATE INDEX IF NOT EXISTS sessions_user_id ON sessions(user_id);

-- +goose Down

PRAGMA foreign_keys = OFF;

CREATE TABLE documents_old AS SELECT * FROM documents;
DROP TABLE documents;
CREATE TABLE documents (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename         TEXT NOT NULL,
    content_type     TEXT NOT NULL,
    data             BLOB NOT NULL,
    state            TEXT NOT NULL DEFAULT 'UPLOADED' CHECK (state IN ('UPLOADED','PARSED','FAILED','SKIPPED','IMPORTED')),
    detected_bank    TEXT,
    extracted_text   TEXT,
    error            TEXT,
    settlement_iban  TEXT,
    transaction_date DATETIME,
    created_at       DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))
);
INSERT INTO documents SELECT id, person_id, filename, content_type, data, state,
    detected_bank, extracted_text, error, settlement_iban, transaction_date, created_at
FROM documents_old;
DROP TABLE documents_old;

CREATE TABLE cash_accounts_old AS SELECT * FROM cash_accounts;
DROP TABLE cash_accounts;
CREATE TABLE cash_accounts (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    currency     TEXT NOT NULL,
    iban         TEXT,
    created_at   DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))
);
INSERT INTO cash_accounts SELECT id, person_id, display_name, currency, iban, created_at FROM cash_accounts_old;
DROP TABLE cash_accounts_old;

CREATE TABLE portfolios_old AS SELECT * FROM portfolios;
DROP TABLE portfolios;
CREATE TABLE portfolios (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name    TEXT NOT NULL,
    cash_account_id TEXT NOT NULL REFERENCES cash_accounts(id) ON DELETE RESTRICT,
    created_at      DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%dT%H:%M:%fZ'))
);
INSERT INTO portfolios SELECT id, person_id, display_name, cash_account_id, created_at FROM portfolios_old;
DROP TABLE portfolios_old;

PRAGMA foreign_keys = ON;

ALTER TABLE sessions DROP COLUMN person_id;
DROP TABLE user_person_access;
DROP TABLE persons;
