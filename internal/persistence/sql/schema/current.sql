-- current.sql is a hand-maintained snapshot of the final schema, used only
-- by sqlc for type inference (see sqlc.yaml's "schema" path). It is never
-- executed — goose applies internal/persistence/sql/migrations/*.sql at
-- runtime, which is the actual source of truth.
--
-- This split exists because sqlc's "sqlite" engine cannot parse
-- `ALTER TABLE ... ALTER COLUMN ... SET NOT NULL` at all (real SQLite has no
-- such statement, so sqlc's grammar has no rule for it) — but lightsql needs
-- exactly that syntax to tighten person_id after backfilling it. Whenever the
-- migrations change a column's final type or nullability, mirror that change
-- here too.

CREATE TABLE users (
    id           TEXT     PRIMARY KEY,
    issuer       TEXT     NOT NULL,
    subject      TEXT     NOT NULL,
    display_name TEXT     NOT NULL,
    created_at   TIMESTAMP NOT NULL,
    UNIQUE (issuer, subject)
);

CREATE TABLE persons (
    id           TEXT      PRIMARY KEY,
    display_name TEXT      NOT NULL,
    created_at   TIMESTAMP NOT NULL
);

CREATE TABLE user_person_access (
    user_id   TEXT NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    person_id TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, person_id)
);

CREATE TABLE cash_accounts (
    id           TEXT     PRIMARY KEY,
    -- Superseded by person_id. Kept only because lightsql cannot DROP
    -- COLUMN; never populated by new rows.
    user_id      TEXT     REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT     NOT NULL,
    currency     TEXT     NOT NULL,
    iban         TEXT,
    created_at   TIMESTAMP NOT NULL,
    person_id    TEXT     NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    -- 'SAVINGS' or 'LOAN'; NULL is a plain cash account. Migration 0013.
    kind              TEXT,
    -- Declared balance (minor units) and when; set = the account is an
    -- estimate, its transactions do not move the balance. Migration 0013.
    estimated_balance INTEGER,
    estimated_at      TIMESTAMP
);

CREATE TABLE portfolios (
    id                TEXT     PRIMARY KEY,
    -- Superseded by person_id. Kept only because lightsql cannot DROP
    -- COLUMN; never populated by new rows.
    user_id           TEXT     REFERENCES users(id) ON DELETE CASCADE,
    display_name      TEXT     NOT NULL,
    cash_account_id   TEXT     NOT NULL REFERENCES cash_accounts(id) ON DELETE RESTRICT,
    created_at        TIMESTAMP NOT NULL,
    person_id         TEXT     NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    -- 'ASSETS' marks the person's "Other assets" portfolio that personal
    -- assets are booked into; NULL for a brokerage portfolio. Migration 0012.
    purpose         TEXT
);

-- The API calls these assets (a security is one kind); the table keeps its
-- historical name, as do security_id columns elsewhere — renaming would
-- churn every foreign key and query for no behaviour.
CREATE TABLE securities (
    id           TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    -- The security this one derives from (a right, option or warrant's
    -- underlying stock); see migrations/0007_underlying_security.sql.
    underlying_security_id TEXT REFERENCES securities(id) ON DELETE SET NULL,
    -- Superseded by asset_kind: its CHECK cannot be dropped and lightsql
    -- cannot rebuild a referenced table. NULL for every new row.
    kind         TEXT CHECK (kind IN ('STOCK', 'ETF', 'FUND', 'BOND', 'OTHER')),
    -- What the asset is: a SecurityKind of the GraphQL schema (STOCK, ETF,
    -- FUND, BOND, PROPERTY, VEHICLE, LOAN, OTHER), null until set. See
    -- migrations/0011_asset_kind.sql.
    asset_kind   TEXT,
    -- The owner of a personal asset (the office room, the car); visible to
    -- that person alone. NULL = shared listed instrument. Migration 0012.
    person_id    TEXT REFERENCES persons(id) ON DELETE CASCADE
);

CREATE TABLE security_identifiers (
    security_id TEXT NOT NULL REFERENCES securities(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL,
    value       TEXT NOT NULL,
    PRIMARY KEY (security_id, kind, value)
);

CREATE TABLE listings (
    id             TEXT PRIMARY KEY,
    security_id    TEXT NOT NULL REFERENCES securities(id) ON DELETE CASCADE,
    exchange       TEXT,
    ticker         TEXT NOT NULL,
    currency       TEXT NOT NULL,
    quote_provider TEXT,
    UNIQUE (security_id, ticker)
);

CREATE TABLE quotes (
    listing_id TEXT    NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    time       TIMESTAMP NOT NULL,
    price      DECIMAL(19,4) NOT NULL,
    PRIMARY KEY (listing_id, time)
);

-- ECB reference rates: how many units of currency one euro buys at time
-- (the reference date, midnight UTC). See migration 0008.
CREATE TABLE fx_rates (
    currency TEXT      NOT NULL,
    time     TIMESTAMP NOT NULL,
    rate     DECIMAL(19,4) NOT NULL,
    PRIMARY KEY (currency, time)
);

CREATE TABLE documents (
    id              TEXT     PRIMARY KEY,
    -- Superseded by person_id. Kept only because lightsql cannot DROP
    -- COLUMN; never populated by new rows.
    user_id         TEXT     REFERENCES users(id) ON DELETE CASCADE,
    filename        TEXT     NOT NULL,
    content_type    TEXT     NOT NULL,
    data            BYTEA    NOT NULL,
    state           TEXT     NOT NULL,
    detected_bank   TEXT,
    extracted_text  TEXT,
    error           TEXT,
    settlement_iban TEXT,
    transaction_date TIMESTAMP,
    created_at      TIMESTAMP NOT NULL,
    person_id       TEXT     NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    -- SHA-256 of data, hex; see migrations/0010_document_hash.sql.
    content_hash    TEXT
);

CREATE TABLE transactions (
    id              TEXT     PRIMARY KEY,
    type            TEXT     NOT NULL,
    time            TIMESTAMP NOT NULL,
    portfolio_id    TEXT     REFERENCES portfolios(id)    ON DELETE CASCADE,
    security_id     TEXT     REFERENCES securities(id)    ON DELETE RESTRICT,
    cash_account_id TEXT     REFERENCES cash_accounts(id) ON DELETE CASCADE,
    units           REAL     NOT NULL,
    price           DECIMAL(19,4),
    fees            INTEGER  NOT NULL,
    taxes           INTEGER  NOT NULL,
    cash_delta      INTEGER  NOT NULL,
    currency        TEXT     NOT NULL,
    source          TEXT     NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    -- The document this transaction was imported from, if any (see
    -- migrations/0004_transaction_document.sql).
    document_id     TEXT     REFERENCES documents(id) ON DELETE SET NULL,
    -- FX bond support (see migrations/0005_fx_rate.sql): NULL price_currency
    -- means Price is denominated in currency (the common case).
    price_currency  TEXT,
    fx_rate         DECIMAL(19,4)
);

CREATE TABLE staged_transactions (
    id            TEXT     PRIMARY KEY,
    document_id   TEXT     NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    type          TEXT     NOT NULL,
    time          TIMESTAMP NOT NULL,
    units         REAL     NOT NULL,
    price         DECIMAL(19,4),
    fees          INTEGER  NOT NULL,
    taxes         INTEGER  NOT NULL,
    cash_delta    INTEGER  NOT NULL,
    currency      TEXT     NOT NULL,
    security_hint TEXT,
    isin          TEXT,
    security_id   TEXT     REFERENCES securities(id) ON DELETE SET NULL,
    created_at    TIMESTAMP NOT NULL,
    price_currency TEXT,
    fx_rate       DECIMAL(19,4),
    -- The kind the parser suggests for a security yet to be created.
    security_kind TEXT
);

CREATE TABLE sessions (
    id         TEXT     PRIMARY KEY,
    user_id    TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    person_id  TEXT REFERENCES persons(id) ON DELETE SET NULL
);
