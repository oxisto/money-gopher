-- +goose Up
CREATE TABLE
    IF NOT EXISTS cash_accounts (
        -- CashAccount is a real-world bank/cash account. Its balance is not
        -- stored but derived as the sum of all transaction cash deltas.
        id TEXT PRIMARY KEY, -- ID is a locally generated identifier.
        user_id TEXT NOT NULL, -- UserID is the owner of the account.
        display_name TEXT NOT NULL, -- DisplayName is the human-readable name of the account.
        currency TEXT NOT NULL, -- Currency is the ISO 4217 code the account is denominated in.
        created_at DATETIME NOT NULL DEFAULT (STRFTIME ('%Y-%m-%dT%H:%M:%fZ')),
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );

CREATE TABLE
    IF NOT EXISTS portfolios (
        -- Portfolio is a collection of security positions belonging to one user.
        -- Cash is deliberately not part of a portfolio; it lives in cash
        -- accounts, and transactions link the two.
        id TEXT PRIMARY KEY, -- ID is a locally generated identifier.
        user_id TEXT NOT NULL, -- UserID is the owner of the portfolio.
        display_name TEXT NOT NULL, -- DisplayName is the human-readable name of the portfolio.
        created_at DATETIME NOT NULL DEFAULT (STRFTIME ('%Y-%m-%dT%H:%M:%fZ')),
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );

CREATE TABLE
    IF NOT EXISTS securities (
        -- Security is a tradable security, shared across all users. External
        -- identifiers such as the ISIN live in security_identifiers.
        id TEXT PRIMARY KEY, -- ID is a locally generated, opaque identifier.
        display_name TEXT NOT NULL -- DisplayName is the human-readable name of the security.
    );

CREATE TABLE
    IF NOT EXISTS security_identifiers (
        -- SecurityIdentifier maps an external identifier (ISIN, WKN, ticker
        -- symbol, ...) to a security, e.g. to match documents during imports.
        security_id TEXT NOT NULL, -- SecurityID is the security this identifier belongs to.
        kind TEXT NOT NULL, -- Kind is the identifier scheme, e.g. 'ISIN' or 'WKN'.
        value TEXT NOT NULL, -- Value is the identifier itself.
        PRIMARY KEY (security_id, kind, value),
        FOREIGN KEY (security_id) REFERENCES securities (id) ON DELETE CASCADE
    );

-- Globally unique identifiers: the same ISIN/WKN cannot point to two
-- securities. Tickers may collide across exchanges, so they are exempt.
CREATE UNIQUE INDEX IF NOT EXISTS idx_security_identifiers_unique ON security_identifiers (kind, value)
WHERE
    kind IN ('ISIN', 'WKN');

CREATE TABLE
    IF NOT EXISTS listings (
        -- Listing is a security listed on a particular exchange, i.e. the
        -- thing that actually has a price.
        id TEXT PRIMARY KEY, -- ID is a locally generated identifier.
        security_id TEXT NOT NULL, -- SecurityID is the security this listing belongs to.
        exchange TEXT, -- Exchange optionally names the exchange, e.g. an MIC code.
        ticker TEXT NOT NULL, -- Ticker is the symbol used on the exchange.
        currency TEXT NOT NULL, -- Currency is the ISO 4217 code quotes are denominated in.
        quote_provider TEXT, -- QuoteProvider is the name of the provider that provides quotes for this listing.
        UNIQUE (security_id, ticker),
        FOREIGN KEY (security_id) REFERENCES securities (id) ON DELETE CASCADE
    );

CREATE TABLE
    IF NOT EXISTS quotes (
        -- Quote is one price observation for a listing; the table is
        -- append-only, the latest quote is simply the newest row.
        listing_id TEXT NOT NULL, -- ListingID is the listing this quote belongs to.
        time DATETIME NOT NULL, -- Time is when the price was observed.
        price INTEGER NOT NULL, -- Price is the price in minor units of the listing currency.
        PRIMARY KEY (listing_id, time),
        FOREIGN KEY (listing_id) REFERENCES listings (id) ON DELETE CASCADE
    );

CREATE TABLE
    IF NOT EXISTS transactions (
        -- Transaction is a typed portfolio/cash event: trades, dividends and
        -- cash movements. Every transaction that moves money names the cash
        -- account it settles against and its signed effect on that account
        -- (cash_delta), so account balances are always derivable.
        --
        -- Monetary values are minor units (e.g. cents) in `currency`; a
        -- transaction is single-currency by design. The share quantity is a
        -- REAL since securities can be held in fractions.
        id TEXT PRIMARY KEY, -- ID is a locally generated identifier.
        type TEXT NOT NULL CHECK (
            type IN (
                'BUY',
                'SELL',
                'DELIVERY_INBOUND',
                'DELIVERY_OUTBOUND',
                'DIVIDEND',
                'INTEREST',
                'DEPOSIT_CASH',
                'WITHDRAW_CASH',
                'ACCOUNT_FEES',
                'TAX_REFUND'
            )
        ), -- Type is the kind of transaction.
        time DATETIME NOT NULL, -- Time is when the transaction happened.
        portfolio_id TEXT, -- PortfolioID is the affected portfolio; NULL for pure cash events.
        security_id TEXT, -- SecurityID is the affected security; NULL for pure cash events.
        cash_account_id TEXT, -- CashAccountID is the settling account; NULL if nothing settles (e.g. deliveries).
        units REAL NOT NULL DEFAULT 0, -- Units is the number of shares involved.
        price INTEGER, -- Price is the price per unit in minor units.
        fees INTEGER NOT NULL DEFAULT 0, -- Fees are the fees in minor units.
        taxes INTEGER NOT NULL DEFAULT 0, -- Taxes are the taxes in minor units.
        cash_delta INTEGER NOT NULL DEFAULT 0, -- CashDelta is the signed change of the cash account balance in minor units.
        currency TEXT NOT NULL, -- Currency is the ISO 4217 code all monetary values of this transaction are in.
        source TEXT NOT NULL DEFAULT 'MANUAL', -- Source records provenance, e.g. 'MANUAL' or an import kind.
        created_at DATETIME NOT NULL DEFAULT (STRFTIME ('%Y-%m-%dT%H:%M:%fZ')),
        CHECK (
            portfolio_id IS NOT NULL
            OR cash_account_id IS NOT NULL
        ),
        FOREIGN KEY (portfolio_id) REFERENCES portfolios (id) ON DELETE CASCADE,
        FOREIGN KEY (security_id) REFERENCES securities (id) ON DELETE RESTRICT,
        FOREIGN KEY (cash_account_id) REFERENCES cash_accounts (id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_time ON transactions (portfolio_id, time);

CREATE INDEX IF NOT EXISTS idx_transactions_cash_account_time ON transactions (cash_account_id, time);

-- +goose Down
DROP INDEX idx_transactions_cash_account_time;

DROP INDEX idx_transactions_portfolio_time;

DROP TABLE transactions;

DROP TABLE quotes;

DROP TABLE listings;

DROP INDEX idx_security_identifiers_unique;

DROP TABLE security_identifiers;

DROP TABLE securities;

DROP TABLE portfolios;

DROP TABLE cash_accounts;
