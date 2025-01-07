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
        type INTEGER NOT NULL, -- Type is the type of the account.
        reference_account_id INTEGER, -- ReferenceAccountID is the ID of the account that this account is related to. For example, if this is a brokerage account, the reference account could be a bank account.
        FOREIGN KEY (reference_account_id) REFERENCES accounts (id) ON DELETE RESTRICT
    );

CREATE TABLE
    IF NOT EXISTS transactions (
        -- Transactions represents a transaction in an account.
        id TEXT PRIMARY KEY, -- ID is the primary identifier for a transaction.
        source_account_id TEXT, -- SourceAccountID is the ID of the account that the transaction originated from.
        destination_account_id TEXT, -- DestinationAccountID is the ID of the account that the transaction is destined for.
        time DATETIME NOT NULL, -- Time is the time that the transaction occurred.
        type INTEGER NOT NULL, -- Type is the type of the transaction. Depending on the type, different fields (source, destination) will be used.
        security_id TEXT, -- SecurityID is the ID of the security that the transaction is related to. Can be empty if the transaction is not related to a security.
        amount REAL NOT NULL, -- Amount is the amount of the transaction.
        price JSONB, -- Price is the price of the transaction.
        fees JSONB, -- Fees is the fees of the transaction.
        taxes JSONB, -- Taxes is the taxes of the transaction.
        FOREIGN KEY (source_account_id) REFERENCES accounts (id) ON DELETE RESTRICT,
        FOREIGN KEY (destination_account_id) REFERENCES accounts (id) ON DELETE RESTRICT
    );

CREATE TABLE
    IF NOT EXISTS portfolio_accounts (
        -- PortfolioAccounts represents the relationship between portfolios and accounts.
        portfolio_id TEXT NOT NULL,
        account_id TEXT NOT NULL,
        FOREIGN KEY (portfolio_id) REFERENCES portfolios (id) ON DELETE RESTRICT,
        FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE RESTRICT,
        PRIMARY KEY (account_id, portfolio_id)
    );

-- +goose Down
DROP TABLE portfolios;

DROP TABLE portfolio_events;

DROP TABLE portfolio_accounts;

DROP TABLE bank_accounts;

DROP TABLE accounts;

DROP TABLE transactions;