-- +goose Up

CREATE TABLE persons (
    id           TEXT      PRIMARY KEY,
    display_name TEXT      NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT (now())
);

-- user_person_access grants a user the ability to view and manage a person's
-- financial data. One user can access many persons; a person can be accessed
-- by many users (e.g. a shared household view).
CREATE TABLE user_person_access (
    user_id   TEXT NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    person_id TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, person_id)
);

-- Create a person for every existing user (same id for a seamless migration).
INSERT INTO persons (id, display_name)
SELECT id, display_name FROM users;

INSERT INTO user_person_access (user_id, person_id)
SELECT id, id FROM users;

-- Add active person to sessions; NULL means "default to first accessible person".
ALTER TABLE sessions ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE SET NULL;

-- lightsql enforces foreign keys unconditionally (no PRAGMA to disable them),
-- so unlike the original SQLite version of this migration, portfolios,
-- cash_accounts and documents are not rebuilt via a _new/rename dance. They
-- gain person_id the way a PostgreSQL migration would: ADD COLUMN, then
-- backfill.
ALTER TABLE portfolios    ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE CASCADE;
ALTER TABLE cash_accounts ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE CASCADE;
ALTER TABLE documents     ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE CASCADE;

UPDATE portfolios    SET person_id = user_id;
UPDATE cash_accounts SET person_id = user_id;
UPDATE documents     SET person_id = user_id;

-- Now that every row has a value, tighten the columns that should be
-- required. lightsql didn't support this when this migration was first
-- written; person_id would otherwise have stayed nullable at the schema
-- level forever even though the application always populates it.
ALTER TABLE portfolios    ALTER COLUMN person_id SET NOT NULL;
ALTER TABLE cash_accounts ALTER COLUMN person_id SET NOT NULL;
ALTER TABLE documents     ALTER COLUMN person_id SET NOT NULL;

-- +goose Down

-- This migration cannot be fully reversed on lightsql today: DROP COLUMN
-- isn't supported, so the person_id foreign keys added to sessions,
-- portfolios, cash_accounts and documents can't be removed, which in turn
-- means `persons` can't be dropped while they still reference it (lightsql
-- always enforces foreign keys, RESTRICT by default). Only the table with no
-- incoming references comes back out.
DROP TABLE user_person_access;
