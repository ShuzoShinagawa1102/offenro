-- +goose Up

INSERT INTO commerce_domain (domain_id, name, status, protocol_version)
VALUES
    ('travel.hotel', 'Travel Hotel', 'ACTIVE', '0.1.0'),
    ('retail.shoes', 'Retail Shoes', 'ACTIVE', '0.1.0')
ON CONFLICT (domain_id) DO NOTHING;

ALTER TABLE cart_item
    ADD COLUMN merchant_offer_ref TEXT;

UPDATE cart_item
SET merchant_offer_ref = offer_id
WHERE merchant_offer_ref IS NULL;

ALTER TABLE cart_item
    ALTER COLUMN merchant_offer_ref SET NOT NULL;

ALTER TABLE purchase_item
    ADD COLUMN merchant_offer_ref TEXT;

UPDATE purchase_item
SET merchant_offer_ref = offer_id
WHERE merchant_offer_ref IS NULL;

ALTER TABLE purchase_item
    ALTER COLUMN merchant_offer_ref SET NOT NULL;

-- +goose Down

ALTER TABLE purchase_item
    DROP COLUMN merchant_offer_ref;

ALTER TABLE cart_item
    DROP COLUMN merchant_offer_ref;

DELETE FROM commerce_domain AS domain
WHERE domain.domain_id IN ('travel.hotel', 'retail.shoes')
  AND NOT EXISTS (
      SELECT 1
      FROM merchant_capability AS capability
      WHERE capability.domain_id = domain.domain_id
  );
