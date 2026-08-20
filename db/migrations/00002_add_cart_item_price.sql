-- +goose Up

ALTER TABLE cart_item
    ADD COLUMN amount BIGINT CHECK (amount >= 0),
    ADD COLUMN currency VARCHAR(3) CHECK (char_length(currency) = 3);

UPDATE cart_item
SET
    amount = (offer_snapshot ->> 'amount')::BIGINT,
    currency = upper(offer_snapshot ->> 'currency')
WHERE offer_snapshot ? 'amount'
  AND offer_snapshot ? 'currency';

ALTER TABLE cart_item
    ALTER COLUMN amount SET NOT NULL,
    ALTER COLUMN currency SET NOT NULL;

-- +goose Down

ALTER TABLE cart_item
    DROP COLUMN currency,
    DROP COLUMN amount;
