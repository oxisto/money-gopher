-- +goose Up
-- An account or contract number that is not an IBAN (a Bausparvertrag's
-- Vertragsnummer, a Depotnummer): what such statements print instead.
ALTER TABLE cash_accounts ADD COLUMN reference TEXT;

-- +goose Down
