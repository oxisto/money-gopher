-- +goose Up

-- DECIMAL(19,4) silently truncated real data: a subscription-rights Kurs
-- of "0,363512" and an FX rate of "1,055432" both have 6 decimal digits,
-- more than the 4 the original migration assumed based on the fund/bond
-- prices seen at the time. Bumping to DECIMAL(19,8) (Portfolio
-- Performance's own precision for the same kind of value) rather than a
-- second minimal bump, on the theory that if 4 wasn't enough, guessing
-- another exact ceiling from a second small sample isn't obviously better
-- than moving to a number with real headroom.
--
-- Same drop+recreate dance as 0003_decimal_prices.sql for the same reason:
-- lightsql has no ALTER COLUMN TYPE, precision changes included. Unlike
-- 0003, transactions now holds real committed rows (it didn't yet when
-- 0003 ran), so it's rebuilt via rename+copy+drop instead of a blind drop —
-- staged_transactions stays a blind drop, since it only ever holds
-- unreviewed drafts regenerated from each document's extracted_text.
DROP TABLE quotes;

CREATE TABLE quotes (
    -- Quote is one price observation for a listing.
    listing_id TEXT    NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    time       TIMESTAMP NOT NULL,
    price      DECIMAL(19,8) NOT NULL,
    PRIMARY KEY (listing_id, time)
);

ALTER TABLE transactions RENAME TO transactions_old_0006;

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
    price           DECIMAL(19,8),
    fees            INTEGER  NOT NULL DEFAULT 0,
    taxes           INTEGER  NOT NULL DEFAULT 0,
    cash_delta      INTEGER  NOT NULL DEFAULT 0,
    currency        TEXT     NOT NULL,
    source          TEXT     NOT NULL DEFAULT 'MANUAL',
    created_at      TIMESTAMP NOT NULL DEFAULT (now()),
    document_id     TEXT     REFERENCES documents(id) ON DELETE SET NULL,
    price_currency  TEXT,
    fx_rate         DECIMAL(19,8),
    CHECK (portfolio_id IS NOT NULL OR cash_account_id IS NOT NULL)
);

INSERT INTO transactions (
    id, type, time, portfolio_id, security_id, cash_account_id, units,
    price, fees, taxes, cash_delta, currency, source, created_at,
    document_id, price_currency, fx_rate
)
SELECT
    id, type, time, portfolio_id, security_id, cash_account_id, units,
    price, fees, taxes, cash_delta, currency, source, created_at,
    document_id, price_currency, fx_rate
FROM transactions_old_0006;

DROP TABLE transactions_old_0006;

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
    price         DECIMAL(19,8),
    fees          INTEGER  NOT NULL DEFAULT 0,
    taxes         INTEGER  NOT NULL DEFAULT 0,
    cash_delta    INTEGER  NOT NULL DEFAULT 0,
    currency      TEXT     NOT NULL,
    security_hint TEXT,
    isin          TEXT,
    security_id   TEXT     REFERENCES securities(id) ON DELETE SET NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT (now()),
    price_currency TEXT,
    fx_rate       DECIMAL(19,8)
);

CREATE INDEX idx_staged_transactions_document
    ON staged_transactions(document_id);

-- Every document whose staged transactions were just discarded needs
-- reparsing; see 0003_decimal_prices.sql for why this is safe (drafts only,
-- regenerated from each document's stored extracted_text).
UPDATE documents SET state = 'UPLOADED' WHERE state = 'PARSED';

-- +goose Down

-- Not reversible: the old DECIMAL(19,4) data was already dropped by the Up
-- migration, and lightsql cannot narrow a column's precision either.
