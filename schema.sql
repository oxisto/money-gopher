CREATE TABLE IF NOT EXISTS securities (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    quote_provider TEXT
);

CREATE TABLE IF NOT EXISTS listed_securities (
    security_id TEXT NOT NULL,
    ticker TEXT NOT NULL,
    currency TEXT NOT NULL,
    latest_quote INTEGER,
    latest_quote_timestamp DATETIME,
    FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE RESTRICT,
    PRIMARY KEY (security_id, ticker)
);