-- +goose Up

CREATE TABLE merchant_fulfillment (
    fulfillment_id TEXT PRIMARY KEY,
    purchase_item_id TEXT NOT NULL UNIQUE REFERENCES purchase_item (purchase_item_id),
    capability_id TEXT NOT NULL REFERENCES merchant_capability (capability_id),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'REJECTED', 'UNKNOWN')),
    idempotency_key TEXT NOT NULL UNIQUE,
    merchant_order_ref TEXT NOT NULL DEFAULT '',
    details_snapshot JSONB NOT NULL,
    response_snapshot JSONB,
    failure_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX merchant_fulfillment_capability_id_idx
    ON merchant_fulfillment (capability_id);

-- +goose Down

DROP TABLE merchant_fulfillment;
