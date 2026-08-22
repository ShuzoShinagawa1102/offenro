-- name: ListCommerceDomains :many
SELECT domain_id, name, status, protocol_version
FROM commerce_domain
ORDER BY domain_id;

-- name: GetCommerceDomain :one
SELECT domain_id, name, status, protocol_version
FROM commerce_domain
WHERE domain_id = $1;
