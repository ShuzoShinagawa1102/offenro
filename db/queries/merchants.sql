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
