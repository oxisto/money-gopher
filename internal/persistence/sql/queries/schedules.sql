-- name: CreateSchedule :one
INSERT INTO
    schedules (
        id,
        person_id,
        display_name,
        type,
        cash_account_id,
        counterpart_account_id,
        security_id,
        recurrence,
        day_of_month,
        start_date,
        end_date,
        books,
        created_at
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetSchedule :one
SELECT
    *
FROM
    schedules
WHERE
    id = ?
    AND person_id = ?;

-- name: ListSchedules :many
SELECT
    *
FROM
    schedules
WHERE
    person_id = ?
ORDER BY
    display_name;

-- name: ListSchedulesByCashAccount :many
-- Schedules that touch the account: landing on it, or paid from it.
SELECT
    *
FROM
    schedules
WHERE
    person_id = ?
    AND (
        cash_account_id = ?
        OR counterpart_account_id = ?
    )
ORDER BY
    display_name;

-- name: ListBookingSchedules :many
-- Every person's schedules that book their occurrences automatically.
SELECT
    *
FROM
    schedules
WHERE
    books = 1
ORDER BY
    person_id,
    display_name;

-- name: UpdateSchedule :one
UPDATE schedules
SET
    display_name = ?,
    counterpart_account_id = ?,
    recurrence = ?,
    day_of_month = ?,
    start_date = ?,
    end_date = ?,
    books = ?
WHERE
    id = ?
    AND person_id = ? RETURNING *;

-- name: DeleteSchedule :execrows
DELETE FROM schedules
WHERE
    id = ?
    AND person_id = ?;

-- name: ListScheduleSteps :many
SELECT
    *
FROM
    schedule_steps
WHERE
    schedule_id = ?
ORDER BY
    from_date;

-- name: ListScheduleStepsByPerson :many
-- All steps of a person's schedules, for expanding them in one go.
SELECT
    schedule_steps.*
FROM
    schedule_steps
    JOIN schedules ON schedules.id = schedule_steps.schedule_id
WHERE
    schedules.person_id = ?
ORDER BY
    schedule_steps.schedule_id,
    schedule_steps.from_date;

-- name: CreateScheduleStep :exec
INSERT INTO
    schedule_steps (schedule_id, from_date, amount)
VALUES
    (?, ?, ?);

-- name: DeleteScheduleSteps :exec
DELETE FROM schedule_steps
WHERE
    schedule_id = ?;
