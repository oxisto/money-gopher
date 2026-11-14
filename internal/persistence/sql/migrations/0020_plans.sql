-- +goose Up
-- A plan: where a person's net worth is heading. The starting amount and
-- the contribution are resolved from the books and the schedules when NULL.
CREATE TABLE plans (
    id              TEXT      PRIMARY KEY,
    person_id       TEXT      NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    display_name    TEXT      NOT NULL,
    -- the first of a month, midnight UTC
    start_date      TIMESTAMP NOT NULL,
    horizon_months  INTEGER   NOT NULL,
    starting_amount INTEGER,
    contribution    INTEGER,
    -- annual cost drag on every growing bucket, in basis points
    cost_bp         INTEGER   NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT (now())
);

-- One row per kind of wealth: the share of the contribution it receives
-- (NULL = split like the current allocation) and its annual return and
-- volatility, in basis points.
CREATE TABLE plan_assumptions (
    plan_id       TEXT    NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    bucket        TEXT    NOT NULL,
    weight_bp     INTEGER,
    return_bp     INTEGER NOT NULL,
    volatility_bp INTEGER NOT NULL,
    PRIMARY KEY (plan_id, bucket)
);

-- +goose Down
