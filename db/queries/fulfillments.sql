-- name: PrepareMerchantFulfillment :execrows
INSERT INTO merchant_fulfillment (
    fulfillment_id,
    purchase_item_id,
    status,
    idempotency_key,
    details_snapshot,
    created_at,
    updated_at
)
VALUES (
    sqlc.arg(fulfillment_id),
    sqlc.arg(purchase_item_id),
    sqlc.arg(status),
    sqlc.arg(idempotency_key),
    sqlc.arg(details_snapshot),
    sqlc.arg(created_at),
    sqlc.arg(updated_at)
)
ON CONFLICT (purchase_item_id) DO NOTHING;

-- name: ListMerchantFulfillments :many
SELECT
    fulfillment.fulfillment_id,
    fulfillment.purchase_item_id,
    fulfillment.status,
    fulfillment.idempotency_key,
    fulfillment.merchant_order_ref,
    fulfillment.details_snapshot,
    fulfillment.response_snapshot,
    fulfillment.failure_code,
    fulfillment.created_at,
    fulfillment.updated_at
FROM merchant_fulfillment AS fulfillment
JOIN purchase_item AS item
  ON item.purchase_item_id = fulfillment.purchase_item_id
WHERE item.purchase_id = $1
ORDER BY fulfillment.purchase_item_id;

-- name: UpdateMerchantFulfillment :execrows
UPDATE merchant_fulfillment
SET
    status = sqlc.arg(status),
    merchant_order_ref = sqlc.narg(merchant_order_ref),
    response_snapshot = sqlc.narg(response_snapshot),
    failure_code = sqlc.narg(failure_code),
    updated_at = sqlc.arg(updated_at)
WHERE fulfillment_id = sqlc.arg(fulfillment_id)
  AND status = sqlc.arg(expected_status);

-- name: UpdatePurchaseStatus :execrows
UPDATE purchase
SET status = sqlc.arg(status)
WHERE purchase_id = sqlc.arg(purchase_id)
  AND status = sqlc.arg(expected_status);
