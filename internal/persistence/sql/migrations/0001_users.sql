-- +goose Up
CREATE TABLE
    IF NOT EXISTS users (
        -- User is a person using money-gopher, provisioned on first OIDC login.
        id TEXT PRIMARY KEY, -- ID is a locally generated identifier.
        issuer TEXT NOT NULL, -- Issuer is the OIDC issuer URL that authenticated the user.
        subject TEXT NOT NULL, -- Subject is the OIDC subject within the issuer.
        display_name TEXT NOT NULL, -- DisplayName is the human-readable name of the user.
        created_at DATETIME NOT NULL DEFAULT (STRFTIME ('%Y-%m-%dT%H:%M:%fZ')),
        UNIQUE (issuer, subject)
    );

-- +goose Down
DROP TABLE users;
