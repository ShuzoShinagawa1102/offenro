-- name: CreateAgent :exec
INSERT INTO agent (agent_id, name, status)
VALUES (sqlc.arg(agent_id), sqlc.arg(name), sqlc.arg(status));

-- name: GetAgent :one
SELECT agent_id, name, status
FROM agent
WHERE agent_id = $1;
