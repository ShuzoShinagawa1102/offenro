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
	ID          string
	CartID      string
	AgentID     string
	BuyerRef    string
	Status      PurchaseStatus
	Items       []PurchaseItem
	TotalAmount int64
	Currency    string
	PurchasedAt time.Time
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
