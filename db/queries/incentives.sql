-- name: CreateIncentiveRule :exec
INSERT INTO incentive_rule (
    incentive_rule_id,
    capability_id,
    reward_type,
    reward_value,
    valid_from,
    valid_to,
    status
)
VALUES (
    sqlc.arg(incentive_rule_id),
    sqlc.arg(capability_id),
    sqlc.arg(reward_type),
    sqlc.arg(reward_value)::NUMERIC(19, 4),
    sqlc.arg(valid_from),
    sqlc.narg(valid_to),
    sqlc.arg(status)
);

-- name: GetIncentiveRule :one
SELECT
    incentive_rule_id,
    capability_id,
    reward_type,
    reward_value::TEXT AS reward_value,
    valid_from,
    valid_to,
    status
FROM incentive_rule
WHERE incentive_rule_id = $1;

-- name: ListIncentiveRules :many
SELECT
    incentive_rule_id,
    capability_id,
    reward_type,
    reward_value::TEXT AS reward_value,
    valid_from,
    valid_to,
    status
FROM incentive_rule
WHERE capability_id = $1
ORDER BY valid_from DESC, incentive_rule_id;

-- name: UpdateIncentiveRule :execrows
UPDATE incentive_rule
SET reward_value = sqlc.arg(reward_value)::NUMERIC(19, 4),
    valid_from = sqlc.arg(valid_from),
    valid_to = sqlc.narg(valid_to),
    status = sqlc.arg(status)
WHERE incentive_rule_id = sqlc.arg(incentive_rule_id);
