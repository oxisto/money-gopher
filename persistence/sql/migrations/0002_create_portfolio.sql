-- +goose Up
CREATE TABLE
    IF NOT EXISTS portfolios (
        id TEXT PRIMARY KEY, -- ID is the primary identifier for a portfolio.     
        display_name TEXT NOT NULL, -- DisplayName is the human-readable name of the portfolio.
        bank_account_id TEXT NOT NULL, -- BankAccountID is the ID of the bank account that holds the portfolio.
        FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id) ON DELETE RESTRICT
    );

CREATE TABLE
    IF NOT EXISTS portfolio_events (
        id TEXT PRIMARY KEY,
        type INTEGER NOT NULL,
        time DATETIME NOT NULL,
        portfolio_id TEXT NOT NULL,
        security_id TEXT NOT NULL,
        amount REAL,
        price INTEGER,
        price_currency TEXT,
        fees INTEGER,
        fees_currency TEXT,
        taxes INTEGER,
        taxes_currency TEXT
    );

CREATE TABLE
    IF NOT EXISTS bank_accounts (
        id TEXT PRIMARY KEY, -- ID is the primary identifier for a bank account.
        display_name TEXT NOT NULL -- DisplayName is the human-readable name of the bank account.
    );

-- +goose Down
DROP TABLE portfolios;

DROP TABLE portfolio_events;

DROP TABLE bank_accounts;