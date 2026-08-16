-- +goose Up

CREATE TABLE commerce_domain (
    domain_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE', 'DEPRECATED')),
    protocol_version TEXT NOT NULL
);

CREATE TABLE merchant (
    merchant_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'ACTIVE', 'SUSPENDED', 'CLOSED'))
);

CREATE TABLE merchant_capability (
    capability_id TEXT PRIMARY KEY,
    merchant_id TEXT NOT NULL REFERENCES merchant (merchant_id),
    domain_id TEXT NOT NULL REFERENCES commerce_domain (domain_id),
    api_base_url TEXT NOT NULL,
    status TEXT NOT NULL CHECK (
        status IN ('PENDING_VERIFICATION', 'ACTIVE', 'VERIFICATION_FAILED', 'SUSPENDED')
    ),
    protocol_version TEXT NOT NULL,
    UNIQUE (merchant_id, domain_id)
);

CREATE TABLE incentive_rule (
    incentive_rule_id TEXT PRIMARY KEY,
    capability_id TEXT NOT NULL REFERENCES merchant_capability (capability_id),
    reward_type TEXT NOT NULL CHECK (reward_type IN ('PERCENTAGE', 'FIXED')),
    reward_value NUMERIC(19, 4) NOT NULL CHECK (reward_value >= 0),
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('DRAFT', 'ACTIVE', 'INACTIVE', 'EXPIRED')),
    CHECK (valid_to IS NULL OR valid_to > valid_from)
);

CREATE TABLE discovery_index_entry (
    index_entry_id TEXT PRIMARY KEY,
    capability_id TEXT NOT NULL REFERENCES merchant_capability (capability_id),
    dimension TEXT NOT NULL,
    value TEXT NOT NULL,
    supply_count INTEGER NOT NULL CHECK (supply_count >= 0),
    indexed_at TIMESTAMPTZ NOT NULL,
    UNIQUE (capability_id, dimension, value)
);

CREATE TABLE agent (
    agent_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'ACTIVE', 'SUSPENDED', 'CLOSED'))
);

CREATE TABLE cart (
    cart_id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL REFERENCES agent (agent_id),
    buyer_ref TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'CHECKED_OUT', 'ABANDONED', 'EXPIRED')),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE cart_item (
    cart_item_id TEXT PRIMARY KEY,
    cart_id TEXT NOT NULL REFERENCES cart (cart_id),
    capability_id TEXT NOT NULL REFERENCES merchant_capability (capability_id),
    offer_id TEXT NOT NULL,
    offer_snapshot JSONB NOT NULL,
    offer_expires_at TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'REMOVED', 'EXPIRED'))
);

CREATE TABLE purchase (
    purchase_id TEXT PRIMARY KEY,
    cart_id TEXT NOT NULL UNIQUE REFERENCES cart (cart_id),
    agent_id TEXT NOT NULL REFERENCES agent (agent_id),
    buyer_ref TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('CREATED', 'CONFIRMED', 'CANCELLED')),
    total_amount NUMERIC(19, 4) NOT NULL CHECK (total_amount >= 0),
    currency VARCHAR(3) NOT NULL CHECK (char_length(currency) = 3),
    purchased_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE purchase_item (
    purchase_item_id TEXT PRIMARY KEY,
    purchase_id TEXT NOT NULL REFERENCES purchase (purchase_id),
    capability_id TEXT NOT NULL REFERENCES merchant_capability (capability_id),
    offer_id TEXT NOT NULL,
    offer_snapshot JSONB NOT NULL,
    amount NUMERIC(19, 4) NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) NOT NULL CHECK (char_length(currency) = 3)
);

CREATE TABLE payment (
    payment_id TEXT PRIMARY KEY,
    purchase_id TEXT NOT NULL REFERENCES purchase (purchase_id),
    provider TEXT NOT NULL,
    provider_payment_id TEXT,
    amount NUMERIC(19, 4) NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) NOT NULL CHECK (char_length(currency) = 3),
    status TEXT NOT NULL CHECK (
        status IN ('PENDING', 'AUTHORIZED', 'CAPTURED', 'FAILED', 'CANCELLED')
    )
);

CREATE TABLE allocation (
    allocation_id TEXT PRIMARY KEY,
    purchase_item_id TEXT NOT NULL REFERENCES purchase_item (purchase_item_id),
    recipient_type TEXT NOT NULL CHECK (recipient_type IN ('MERCHANT', 'AGENT', 'PROTOCOL')),
    recipient_id TEXT NOT NULL,
    amount NUMERIC(19, 4) NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) NOT NULL CHECK (char_length(currency) = 3),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'TRANSFERRED', 'REVERSED'))
);

CREATE TABLE transfer (
    transfer_id TEXT PRIMARY KEY,
    allocation_id TEXT NOT NULL REFERENCES allocation (allocation_id),
    provider_transfer_id TEXT,
    amount NUMERIC(19, 4) NOT NULL CHECK (amount >= 0),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED', 'REVERSED'))
);

CREATE TABLE refund (
    refund_id TEXT PRIMARY KEY,
    payment_id TEXT NOT NULL REFERENCES payment (payment_id),
    provider_refund_id TEXT,
    amount NUMERIC(19, 4) NOT NULL CHECK (amount >= 0),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED', 'CANCELLED'))
);

CREATE INDEX merchant_capability_domain_id_idx
    ON merchant_capability (domain_id);
CREATE INDEX incentive_rule_capability_id_idx
    ON incentive_rule (capability_id);
CREATE INDEX discovery_index_entry_lookup_idx
    ON discovery_index_entry (dimension, value);
CREATE INDEX cart_agent_id_idx
    ON cart (agent_id);
CREATE INDEX cart_item_cart_id_idx
    ON cart_item (cart_id);
CREATE INDEX cart_item_capability_id_idx
    ON cart_item (capability_id);
CREATE INDEX purchase_agent_id_idx
    ON purchase (agent_id);
CREATE INDEX purchase_item_purchase_id_idx
    ON purchase_item (purchase_id);
CREATE INDEX purchase_item_capability_id_idx
    ON purchase_item (capability_id);
CREATE INDEX payment_purchase_id_idx
    ON payment (purchase_id);
CREATE INDEX allocation_purchase_item_id_idx
    ON allocation (purchase_item_id);
CREATE INDEX transfer_allocation_id_idx
    ON transfer (allocation_id);
CREATE INDEX refund_payment_id_idx
    ON refund (payment_id);

-- +goose Down

DROP TABLE refund;
DROP TABLE transfer;
DROP TABLE allocation;
DROP TABLE payment;
DROP TABLE purchase_item;
DROP TABLE purchase;
DROP TABLE cart_item;
DROP TABLE cart;
DROP TABLE agent;
DROP TABLE discovery_index_entry;
DROP TABLE incentive_rule;
DROP TABLE merchant_capability;
DROP TABLE merchant;
DROP TABLE commerce_domain;
