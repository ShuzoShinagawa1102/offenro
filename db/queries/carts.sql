-- name: CreateCart :execrows
INSERT INTO cart (
    cart_id,
    agent_id,
    buyer_ref,
    status,
    created_at,
    updated_at
)
SELECT
    sqlc.arg(cart_id),
    registered_agent.agent_id,
    sqlc.arg(buyer_ref),
    sqlc.arg(status),
    sqlc.arg(created_at),
    sqlc.arg(updated_at)
FROM agent AS registered_agent
WHERE registered_agent.agent_id = sqlc.arg(agent_id)
  AND registered_agent.status = 'ACTIVE';

-- name: GetCart :one
SELECT cart_id, agent_id, buyer_ref, status, created_at, updated_at
FROM cart
WHERE cart_id = $1;

-- name: ListCartItems :many
SELECT
    item.cart_item_id,
    item.cart_id,
    item.capability_id,
    capability.merchant_id,
    capability.domain_id,
    item.offer_id,
    item.offer_snapshot,
    item.amount,
    item.currency,
    item.offer_expires_at,
    item.status
FROM cart_item AS item
JOIN merchant_capability AS capability
  ON capability.capability_id = item.capability_id
WHERE item.cart_id = $1
ORDER BY item.cart_item_id;

-- name: UpdateCart :execrows
UPDATE cart
SET buyer_ref = sqlc.arg(buyer_ref),
    status = sqlc.arg(status),
    updated_at = sqlc.arg(updated_at)
WHERE cart_id = sqlc.arg(cart_id)
  AND status = 'ACTIVE'
  AND updated_at = sqlc.arg(expected_updated_at);

-- name: TouchActiveCart :execrows
UPDATE cart
SET updated_at = sqlc.arg(updated_at)
WHERE cart_id = sqlc.arg(cart_id)
  AND status = 'ACTIVE'
  AND updated_at = sqlc.arg(expected_updated_at);

-- name: CreateCartItem :one
INSERT INTO cart_item (
    cart_item_id,
    cart_id,
    capability_id,
    offer_id,
    offer_snapshot,
    offer_expires_at,
    status,
    amount,
    currency
)
SELECT
    sqlc.arg(cart_item_id),
    sqlc.arg(cart_id),
    capability.capability_id,
    sqlc.arg(offer_id),
    sqlc.arg(offer_snapshot),
    sqlc.narg(offer_expires_at),
    sqlc.arg(status),
    sqlc.arg(amount),
    sqlc.arg(currency)
FROM merchant_capability AS capability
JOIN merchant ON merchant.merchant_id = capability.merchant_id
JOIN commerce_domain AS domain ON domain.domain_id = capability.domain_id
WHERE capability.merchant_id = sqlc.arg(merchant_id)
  AND capability.domain_id = sqlc.arg(domain_id)
  AND capability.status = 'ACTIVE'
  AND merchant.status = 'ACTIVE'
  AND domain.status = 'ACTIVE'
RETURNING capability_id;

-- name: UpdateCartItem :execrows
UPDATE cart_item
SET offer_snapshot = sqlc.arg(offer_snapshot),
    offer_expires_at = sqlc.narg(offer_expires_at),
    status = sqlc.arg(status),
    amount = sqlc.arg(amount),
    currency = sqlc.arg(currency)
WHERE cart_item_id = sqlc.arg(cart_item_id)
  AND cart_id = sqlc.arg(cart_id)
  AND status = 'ACTIVE';

-- name: MarkCartCheckedOut :execrows
UPDATE cart
SET status = 'CHECKED_OUT',
    updated_at = sqlc.arg(updated_at)
WHERE cart_id = sqlc.arg(cart_id)
  AND status = 'ACTIVE'
  AND updated_at = sqlc.arg(expected_updated_at);
