-- +goose Up

-- A security is one kind of asset; the office room rented out or the car
-- are others. The kinds therefore grow beyond stock/ETF/fund/bond, but
-- securities.kind carries a CHECK that lists them, and lightsql can neither
-- drop a constraint nor rebuild a table that other tables reference. So the
-- kind moves to a column without one; kind stays as it is, NULL for every
-- new row, and is never read again. Validation lives in the GraphQL enum.
ALTER TABLE securities ADD COLUMN asset_kind TEXT;
UPDATE securities SET asset_kind = kind WHERE kind IS NOT NULL;

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
