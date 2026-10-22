-- +goose Up

-- What the account is: 'SAVINGS' (a Bausparvertrag, a Tagesgeld used as
-- savings) or 'LOAN' (a liability whose balance runs negative). NULL is a
-- plain cash account, as every row so far. No CHECK: the GraphQL
-- AccountKind enum validates, since lightsql cannot drop a constraint.
ALTER TABLE cash_accounts ADD COLUMN kind TEXT;

-- Set when the balance is declared rather than tracked ("about 1,500 on
-- the Giro"): the declared balance in minor units and when it was
-- declared. Transactions on such an account still exist, for the
-- portfolio that settles against it, but do not move its balance.
ALTER TABLE cash_accounts ADD COLUMN estimated_balance INTEGER;
ALTER TABLE cash_accounts ADD COLUMN estimated_at TIMESTAMP;

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
