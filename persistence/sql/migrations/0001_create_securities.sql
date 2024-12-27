-- +goose Up
CREATE TABLE
    IF NOT EXISTS securities (
        -- Security represents a security that can be traded on an exchange.
        id TEXT PRIMARY KEY, -- ID is the primary identifier for a security.     
        display_name TEXT NOT NULL, -- DisplayName is the human-readable name of the security.
        quote_provider TEXT -- QuoteProvider is the name of the provider that provides quotes for this security.
    );

CREATE TABLE
    IF NOT EXISTS listed_securities (
        -- ListedSecurity represents a security that is listed on a particular exchange.
        security_id TEXT NOT NULL, -- SecurityID is the ID of the security.
        ticker TEXT NOT NULL, -- Ticker is the symbol used to identify the security on the exchange.
        currency TEXT NOT NULL, -- Currency is the currency in which the security is traded.
        latest_quote INTEGER, -- LatestQuote is the latest quote for the security.
        latest_quote_timestamp DATETIME, -- LatestQuoteTimestamp is the timestamp of the latest quote.
        FOREIGN KEY (security_id) REFERENCES securities (id) ON DELETE RESTRICT,
        PRIMARY KEY (security_id, ticker)
    );

-- +goose Down
DROP TABLE securities;
DROP TABLE listed_securities;
