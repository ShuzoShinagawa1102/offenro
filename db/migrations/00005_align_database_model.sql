-- +goose Up

-- Merchant Capability is already determined by purchase_item.capability_id.
-- Keeping the same foreign key on merchant_fulfillment would allow contradictory rows.
DROP INDEX merchant_fulfillment_capability_id_idx;

ALTER TABLE merchant_fulfillment
    DROP COLUMN capability_id;

UPDATE merchant_fulfillment
SET merchant_order_ref = NULL
WHERE merchant_order_ref = '';

UPDATE merchant_fulfillment
SET failure_code = NULL
WHERE failure_code = '';

ALTER TABLE merchant_fulfillment
    ALTER COLUMN merchant_order_ref DROP NOT NULL,
    ALTER COLUMN merchant_order_ref DROP DEFAULT,
    ALTER COLUMN failure_code DROP NOT NULL,
    ALTER COLUMN failure_code DROP DEFAULT;

-- Offenro represents money as an integer count of the currency's minor unit.
-- Abort instead of silently rounding legacy fractional values.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM purchase WHERE total_amount <> trunc(total_amount))
        OR EXISTS (SELECT 1 FROM purchase_item WHERE amount <> trunc(amount))
        OR EXISTS (SELECT 1 FROM payment WHERE amount <> trunc(amount))
        OR EXISTS (SELECT 1 FROM allocation WHERE amount <> trunc(amount))
        OR EXISTS (SELECT 1 FROM transfer WHERE amount <> trunc(amount))
        OR EXISTS (SELECT 1 FROM refund WHERE amount <> trunc(amount)) THEN
        RAISE EXCEPTION 'fractional monetary values must be converted to minor units before migration';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE purchase
    ALTER COLUMN total_amount TYPE BIGINT USING total_amount::BIGINT;

ALTER TABLE purchase_item
    ALTER COLUMN amount TYPE BIGINT USING amount::BIGINT;

ALTER TABLE payment
    ALTER COLUMN amount TYPE BIGINT USING amount::BIGINT;

ALTER TABLE allocation
    ALTER COLUMN amount TYPE BIGINT USING amount::BIGINT;

ALTER TABLE transfer
    ALTER COLUMN amount TYPE BIGINT USING amount::BIGINT;

ALTER TABLE refund
    ALTER COLUMN amount TYPE BIGINT USING amount::BIGINT;

-- +goose Down

ALTER TABLE refund
    ALTER COLUMN amount TYPE NUMERIC(19, 4) USING amount::NUMERIC(19, 4);

ALTER TABLE transfer
    ALTER COLUMN amount TYPE NUMERIC(19, 4) USING amount::NUMERIC(19, 4);

ALTER TABLE allocation
    ALTER COLUMN amount TYPE NUMERIC(19, 4) USING amount::NUMERIC(19, 4);

ALTER TABLE payment
    ALTER COLUMN amount TYPE NUMERIC(19, 4) USING amount::NUMERIC(19, 4);

ALTER TABLE purchase_item
    ALTER COLUMN amount TYPE NUMERIC(19, 4) USING amount::NUMERIC(19, 4);

ALTER TABLE purchase
    ALTER COLUMN total_amount TYPE NUMERIC(19, 4) USING total_amount::NUMERIC(19, 4);

ALTER TABLE merchant_fulfillment
    ADD COLUMN capability_id TEXT;

UPDATE merchant_fulfillment AS fulfillment
SET capability_id = item.capability_id
FROM purchase_item AS item
WHERE item.purchase_item_id = fulfillment.purchase_item_id;

ALTER TABLE merchant_fulfillment
    ALTER COLUMN capability_id SET NOT NULL,
    ADD CONSTRAINT merchant_fulfillment_capability_id_fkey
        FOREIGN KEY (capability_id) REFERENCES merchant_capability (capability_id);

CREATE INDEX merchant_fulfillment_capability_id_idx
    ON merchant_fulfillment (capability_id);

UPDATE merchant_fulfillment
SET merchant_order_ref = ''
WHERE merchant_order_ref IS NULL;

UPDATE merchant_fulfillment
SET failure_code = ''
WHERE failure_code IS NULL;

ALTER TABLE merchant_fulfillment
    ALTER COLUMN merchant_order_ref SET DEFAULT '',
    ALTER COLUMN merchant_order_ref SET NOT NULL,
    ALTER COLUMN failure_code SET DEFAULT '',
    ALTER COLUMN failure_code SET NOT NULL;
