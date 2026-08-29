package model

import (
	"encoding/json"
	"time"
)

type CartStatus string

const (
	CartStatusActive     CartStatus = "ACTIVE"
	CartStatusCheckedOut CartStatus = "CHECKED_OUT"
	CartStatusAbandoned  CartStatus = "ABANDONED"
	CartStatusExpired    CartStatus = "EXPIRED"
)

type CartItemStatus string

const (
	CartItemStatusActive  CartItemStatus = "ACTIVE"
	CartItemStatusRemoved CartItemStatus = "REMOVED"
	CartItemStatusExpired CartItemStatus = "EXPIRED"
)

type PurchaseStatus string

const (
	PurchaseStatusCreated   PurchaseStatus = "CREATED"
	PurchaseStatusConfirmed PurchaseStatus = "CONFIRMED"
	PurchaseStatusCancelled PurchaseStatus = "CANCELLED"
)

type Cart struct {
	ID        string
	AgentID   string
	BuyerRef  string
	Status    CartStatus
	Items     []CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CartItem struct {
	ID               string
	CartID           string
	CapabilityID     string
	MerchantID       MerchantID
	Domain           Domain
	OfferID          string
	MerchantOfferRef string
	OfferSnapshot    json.RawMessage
	Amount           int64
	Currency         string
	OfferExpiresAt   *time.Time
	Status           CartItemStatus
}

type Purchase struct {
	ID           string
	CartID       string
	AgentID      string
	BuyerRef     string
	Status       PurchaseStatus
	Items        []PurchaseItem
	Fulfillments []MerchantFulfillment
	TotalAmount  int64
	Currency     string
	PurchasedAt  time.Time
}

type PurchaseItem struct {
	ID               string
	PurchaseID       string
	CapabilityID     string
	MerchantID       MerchantID
	Domain           Domain
	OfferID          string
	MerchantOfferRef string
	OfferSnapshot    json.RawMessage
	Amount           int64
	Currency         string
}

type MerchantFulfillmentStatus string

const (
	MerchantFulfillmentStatusPending   MerchantFulfillmentStatus = "PENDING"
	MerchantFulfillmentStatusConfirmed MerchantFulfillmentStatus = "CONFIRMED"
	MerchantFulfillmentStatusRejected  MerchantFulfillmentStatus = "REJECTED"
	MerchantFulfillmentStatusUnknown   MerchantFulfillmentStatus = "UNKNOWN"
)

// MerchantFulfillment tracks the merchant-side order or reservation for one PurchaseItem.
// DetailsSnapshot is opaque to Protocol Core and interpreted only by its domain extension.
type MerchantFulfillment struct {
	ID               string
	PurchaseItemID   string
	Status           MerchantFulfillmentStatus
	IdempotencyKey   string
	MerchantOrderRef string
	DetailsSnapshot  json.RawMessage
	ResponseSnapshot json.RawMessage
	FailureCode      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
