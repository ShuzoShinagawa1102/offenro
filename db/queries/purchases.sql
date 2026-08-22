-- name: CreatePurchase :exec
INSERT INTO purchase (
    purchase_id,
    cart_id,
    agent_id,
    buyer_ref,
    status,
    total_amount,
    currency,
    purchased_at
)
VALUES (
    sqlc.arg(purchase_id),
    sqlc.arg(cart_id),
    sqlc.arg(agent_id),
    sqlc.arg(buyer_ref),
    sqlc.arg(status),
    sqlc.arg(total_amount)::BIGINT,
    sqlc.arg(currency),
    sqlc.arg(purchased_at)
);

-- name: CreatePurchaseItem :exec
INSERT INTO purchase_item (
    purchase_item_id,
    purchase_id,
    capability_id,
    offer_id,
    merchant_offer_ref,
    offer_snapshot,
    amount,
    currency
)
VALUES (
    sqlc.arg(purchase_item_id),
    sqlc.arg(purchase_id),
    sqlc.arg(capability_id),
    sqlc.arg(offer_id),
    sqlc.arg(merchant_offer_ref),
    sqlc.arg(offer_snapshot),
    sqlc.arg(amount)::BIGINT,
    sqlc.arg(currency)
);

-- name: GetPurchase :one
SELECT
    purchase_id,
    cart_id,
    agent_id,
    buyer_ref,
    status,
    total_amount::BIGINT AS total_amount,
    currency,
    purchased_at
FROM purchase
WHERE purchase_id = $1;

-- name: ListPurchaseItems :many
SELECT
    item.purchase_item_id,
    item.purchase_id,
    item.capability_id,
    capability.merchant_id,
    capability.domain_id,
    item.offer_id,
    item.merchant_offer_ref,
    item.offer_snapshot,
    item.amount::BIGINT AS amount,
    item.currency
FROM purchase_item AS item
JOIN merchant_capability AS capability
  ON capability.capability_id = item.capability_id
WHERE item.purchase_id = $1
ORDER BY item.purchase_item_id;
