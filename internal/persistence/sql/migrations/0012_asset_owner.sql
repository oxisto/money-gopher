-- +goose Up

-- A security is one kind of asset. A personal asset — the office room
-- rented out, the car, a loan given — is a securities row with an owner:
-- it is visible to and editable by that person alone. NULL means shared,
-- as every row so far (listed instruments).
ALTER TABLE securities ADD COLUMN person_id TEXT REFERENCES persons(id) ON DELETE CASCADE;

-- Personal assets are held in one portfolio per person, "Other assets",
-- created on first use. It is found by purpose, not by its name, so that
-- renaming it does not conjure a second one. NULL for brokerage portfolios.
ALTER TABLE portfolios ADD COLUMN purpose TEXT;

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
