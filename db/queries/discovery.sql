-- name: DeleteDiscoveryIndexByDomain :exec
DELETE FROM discovery_index_entry AS entry
USING merchant_capability AS capability
WHERE entry.capability_id = capability.capability_id
  AND capability.domain_id = $1;

-- name: CreateDiscoveryIndexEntry :execrows
INSERT INTO discovery_index_entry (
    index_entry_id,
    capability_id,
    dimension,
    value,
    supply_count,
    indexed_at
)
SELECT
    sqlc.arg(index_entry_id),
    capability_id,
    sqlc.arg(dimension),
    sqlc.arg(value),
    sqlc.arg(supply_count),
    sqlc.arg(indexed_at)
FROM merchant_capability
WHERE merchant_id = sqlc.arg(merchant_id)
  AND domain_id = sqlc.arg(domain_id)
  AND status = 'ACTIVE';

-- name: FindDiscoveryIndexEntries :many
SELECT
    capability.domain_id,
    entry.dimension,
    entry.value,
    capability.merchant_id,
    entry.supply_count,
    entry.indexed_at
FROM discovery_index_entry AS entry
JOIN merchant_capability AS capability
  ON capability.capability_id = entry.capability_id
WHERE capability.domain_id = sqlc.arg(domain_id)
  AND capability.status = 'ACTIVE'
  AND entry.dimension = sqlc.arg(dimension)
  AND entry.value = sqlc.arg(value)
ORDER BY entry.supply_count DESC, capability.merchant_id
LIMIT sqlc.arg(result_limit);
