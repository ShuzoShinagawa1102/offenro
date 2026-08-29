package fulfillment

import (
	"context"
	"encoding/json"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type Repository interface {
	FindPurchase(ctx context.Context, purchaseID string) (model.Purchase, error)
	PrepareFulfillments(ctx context.Context, values []model.MerchantFulfillment) error
	ListFulfillments(ctx context.Context, purchaseID string) ([]model.MerchantFulfillment, error)
	UpdateFulfillment(ctx context.Context, value model.MerchantFulfillment, expectedStatus model.MerchantFulfillmentStatus) error
	UpdatePurchaseStatus(ctx context.Context, purchaseID string, expectedStatus, status model.PurchaseStatus) error
}

// DomainFulfiller owns domain-specific validation and merchant API calls.
// Protocol Core deliberately treats Details as opaque JSON.
type DomainFulfiller interface {
	Domain() model.Domain
	ValidateDetails(details json.RawMessage) error
	Submit(ctx context.Context, capability model.MerchantCapability, request MerchantRequest) (MerchantResult, error)
	Resolve(ctx context.Context, capability model.MerchantCapability, idempotencyKey string) (MerchantResult, error)
}

type FulfillerRegistry interface {
	Fulfiller(domain model.Domain) (DomainFulfiller, bool)
}

type MerchantRequest struct {
	PurchaseID       string
	PurchaseItemID   string
	BuyerRef         string
	MerchantOfferRef string
	IdempotencyKey   string
	Details          json.RawMessage
}

type MerchantResult struct {
	Status           model.MerchantFulfillmentStatus
	MerchantOrderRef string
	FailureCode      string
	Snapshot         json.RawMessage
}

type ConfirmInput struct {
	PurchaseID string
	Details    map[string]json.RawMessage
}

type UseCases interface {
	Confirm(ctx context.Context, input ConfirmInput) (model.Purchase, error)
}
