-- +goose Up
CREATE TABLE
    IF NOT EXISTS portfolios (
        -- Portfolios represent a collection of securities and other positions
        -- held by a user.
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
        amount REAL NOT NULL,
        price JSONB,
        fees JSONB,
        taxes JSONB
    );

CREATE TABLE
    IF NOT EXISTS bank_accounts (
        id TEXT PRIMARY KEY, -- ID is the primary identifier for a bank account.
        display_name TEXT NOT NULL -- DisplayName is the human-readable name of the bank account.
    );

CREATE TABLE
    IF NOT EXISTS accounts (
        -- Accounts represents an account, such as a brokerage account or a bank
        -- account which comprise a portfolio.
        id TEXT PRIMARY KEY, -- ID is the primary identifier for a brokerage account.
        display_name TEXT NOT NULL, -- DisplayName is the human-readable name of the brokerage account.
        type INTEGER NOT NULL -- Type is the type of the account.
    );

-- +goose Down
DROP TABLE portfolios;

DROP TABLE portfolio_events;

DROP TABLE bank_accounts;

DROP TABLE accounts;