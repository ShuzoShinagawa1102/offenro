-- name: FindActiveMerchantCapabilitiesByDomain :many
SELECT
    capability.capability_id,
    merchant.merchant_id,
    capability.api_base_url,
    capability.domain_id
FROM merchant_capability AS capability
JOIN merchant ON merchant.merchant_id = capability.merchant_id
JOIN commerce_domain AS domain ON domain.domain_id = capability.domain_id
WHERE capability.domain_id = $1
  AND capability.status = 'ACTIVE'
  AND merchant.status = 'ACTIVE'
  AND domain.status = 'ACTIVE'
ORDER BY merchant.merchant_id;

-- name: GetActiveMerchantCapability :one
SELECT
    capability.capability_id,
    merchant.merchant_id,
    merchant.name AS merchant_name,
    merchant.status AS merchant_status,
    capability.api_base_url,
    capability.domain_id,
    capability.status AS capability_status,
    capability.protocol_version
FROM merchant_capability AS capability
JOIN merchant ON merchant.merchant_id = capability.merchant_id
JOIN commerce_domain AS domain ON domain.domain_id = capability.domain_id
WHERE capability.capability_id = $1
  AND capability.status = 'ACTIVE'
  AND merchant.status = 'ACTIVE'
  AND domain.status = 'ACTIVE';

-- name: CreateMerchant :exec
INSERT INTO merchant (merchant_id, name, status)
VALUES (sqlc.arg(merchant_id), sqlc.arg(name), sqlc.arg(status));

-- name: GetMerchant :one
SELECT merchant_id, name, status
FROM merchant
WHERE merchant_id = $1;

-- name: UpdateMerchant :execrows
UPDATE merchant
SET name = sqlc.arg(name),
    status = sqlc.arg(status)
WHERE merchant_id = sqlc.arg(merchant_id);

-- name: CreateMerchantCapability :exec
INSERT INTO merchant_capability (
    capability_id,
    merchant_id,
    domain_id,
    api_base_url,
    status,
    protocol_version
)
VALUES (
    sqlc.arg(capability_id),
    sqlc.arg(merchant_id),
    sqlc.arg(domain_id),
    sqlc.arg(api_base_url),
    sqlc.arg(status),
    sqlc.arg(protocol_version)
);

-- name: GetMerchantCapability :one
SELECT
    capability.capability_id,
    merchant.merchant_id,
    merchant.name AS merchant_name,
    merchant.status AS merchant_status,
    capability.domain_id,
    capability.api_base_url,
    capability.status AS capability_status,
    capability.protocol_version
FROM merchant_capability AS capability
JOIN merchant ON merchant.merchant_id = capability.merchant_id
WHERE capability.capability_id = $1;

-- name: ListMerchantCapabilities :many
SELECT
    capability.capability_id,
    merchant.merchant_id,
    merchant.name AS merchant_name,
    merchant.status AS merchant_status,
    capability.domain_id,
    capability.api_base_url,
    capability.status AS capability_status,
    capability.protocol_version
FROM merchant_capability AS capability
JOIN merchant ON merchant.merchant_id = capability.merchant_id
WHERE capability.merchant_id = $1
ORDER BY capability.domain_id;

-- name: UpdateMerchantCapability :execrows
UPDATE merchant_capability
SET api_base_url = sqlc.arg(api_base_url),
    status = sqlc.arg(status),
    protocol_version = sqlc.arg(protocol_version)
WHERE capability_id = sqlc.arg(capability_id);

-- name: SetMerchantCapabilityStatus :execrows
UPDATE merchant_capability
SET status = sqlc.arg(status)
WHERE capability_id = sqlc.arg(capability_id);

-- name: ActivateMerchantForCapability :execrows
UPDATE merchant
SET status = 'ACTIVE'
WHERE merchant_id = (
    SELECT merchant_id
    FROM merchant_capability
    WHERE capability_id = sqlc.arg(capability_id)
)
  AND status IN ('PENDING', 'ACTIVE');
