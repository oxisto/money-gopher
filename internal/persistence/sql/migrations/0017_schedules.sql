-- +goose Up
-- Recurring cash flows: a salary, a Sparrate, a loan's interest, the rent an
-- office room brings in. A schedule lands on cash_account_id; a TRANSFER
-- is paid from counterpart_account_id; a DIVIDEND names its security.
-- books = 1 means due occurrences are written as real transactions.
CREATE TABLE schedules (
    id                     TEXT      PRIMARY KEY,
    person_id              TEXT      NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    display_name           TEXT      NOT NULL,
    -- DEPOSIT_CASH, WITHDRAW_CASH, INTEREST, ACCOUNT_FEES, DIVIDEND, TRANSFER
    type                   TEXT      NOT NULL,
    cash_account_id        TEXT      NOT NULL REFERENCES cash_accounts(id) ON DELETE CASCADE,
    counterpart_account_id TEXT      REFERENCES cash_accounts(id) ON DELETE CASCADE,
    security_id            TEXT      REFERENCES securities(id) ON DELETE CASCADE,
    -- MONTHLY, QUARTERLY, YEARLY
    recurrence             TEXT      NOT NULL,
    day_of_month           INTEGER   NOT NULL,
    start_date             TIMESTAMP NOT NULL,
    end_date               TIMESTAMP,
    books                  INTEGER   NOT NULL DEFAULT 0,
    created_at             TIMESTAMP NOT NULL DEFAULT (now())
);

-- The amount in force from a date on; the first step gates the schedule.
CREATE TABLE schedule_steps (
    schedule_id TEXT      NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    from_date   TIMESTAMP NOT NULL,
    amount      INTEGER   NOT NULL,
    PRIMARY KEY (schedule_id, from_date)
);

-- A transaction a schedule booked: which one, and for which occurrence.
-- document_id stays NULL until a statement confirms it.
ALTER TABLE transactions ADD COLUMN schedule_id TEXT;
ALTER TABLE transactions ADD COLUMN scheduled_for TIMESTAMP;
CREATE INDEX idx_transactions_schedule ON transactions(schedule_id, scheduled_for);

-- +goose Down
