-- name: CreatePlan :one
INSERT INTO
    plans (
        id,
        person_id,
        display_name,
        start_date,
        horizon_months,
        starting_amount,
        contribution,
        cost_bp,
        created_at
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetPlan :one
SELECT
    *
FROM
    plans
WHERE
    id = ?
    AND person_id = ?;

-- name: ListPlans :many
SELECT
    *
FROM
    plans
WHERE
    person_id = ?
ORDER BY
    created_at,
    display_name;

-- name: UpdatePlan :one
UPDATE plans
SET
    display_name = ?,
    start_date = ?,
    horizon_months = ?,
    starting_amount = ?,
    contribution = ?,
    cost_bp = ?
WHERE
    id = ?
    AND person_id = ? RETURNING *;

-- name: DeletePlan :execrows
DELETE FROM plans
WHERE
    id = ?
    AND person_id = ?;

-- name: ListPlanAssumptions :many
SELECT
    *
FROM
    plan_assumptions
WHERE
    plan_id = ?
ORDER BY
    bucket;

-- name: CreatePlanAssumption :exec
INSERT INTO
    plan_assumptions (plan_id, bucket, weight_bp, return_bp, volatility_bp)
VALUES
    (?, ?, ?, ?, ?);

-- name: DeletePlanAssumptions :exec
DELETE FROM plan_assumptions
WHERE
    plan_id = ?;
