package fulfillment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type Service struct {
	repository Repository
	merchants  merchant.Registry
	fulfillers FulfillerRegistry
}

var _ UseCases = (*Service)(nil)

func NewService(repository Repository, merchants merchant.Registry, fulfillers FulfillerRegistry) *Service {
	return &Service{repository: repository, merchants: merchants, fulfillers: fulfillers}
}

func (s *Service) Confirm(ctx context.Context, input ConfirmInput) (model.Purchase, error) {
	if strings.TrimSpace(input.PurchaseID) == "" {
		return model.Purchase{}, fmt.Errorf("%w: purchase_id is required", ErrInvalidInput)
	}
	purchase, err := s.repository.FindPurchase(ctx, input.PurchaseID)
	if err != nil {
		return model.Purchase{}, err
	}
	if purchase.Status == model.PurchaseStatusConfirmed || purchase.Status == model.PurchaseStatusCancelled {
		return purchase, nil
	}
	if purchase.Status != model.PurchaseStatusCreated {
		return model.Purchase{}, fmt.Errorf("%w: purchase cannot be confirmed from %s", ErrConflict, purchase.Status)
	}

	existing, err := s.repository.ListFulfillments(ctx, purchase.ID)
	if err != nil {
		return model.Purchase{}, err
	}
	byItem := make(map[string]model.MerchantFulfillment, len(existing))
	for _, value := range existing {
		byItem[value.PurchaseItemID] = value
	}
	items := make(map[string]model.PurchaseItem, len(purchase.Items))
	for _, item := range purchase.Items {
		items[item.ID] = item
	}
	for itemID := range input.Details {
		if _, ok := items[itemID]; !ok {
			return model.Purchase{}, fmt.Errorf("%w: unknown purchase_item_id %s", ErrInvalidInput, itemID)
		}
	}

	now := time.Now().UTC()
	prepared := make([]model.MerchantFulfillment, 0, len(purchase.Items)-len(existing))
	for _, item := range purchase.Items {
		if _, ok := byItem[item.ID]; ok {
			continue
		}
		details, ok := input.Details[item.ID]
		if !ok {
			return model.Purchase{}, fmt.Errorf("%w: fulfillment details are required for purchase_item_id %s", ErrInvalidInput, item.ID)
		}
		fulfiller, ok := s.fulfillers.Fulfiller(item.Domain)
		if !ok {
			return model.Purchase{}, fmt.Errorf("%w: fulfillment is unavailable for domain %s", ErrConflict, item.Domain)
		}
		if err := fulfiller.ValidateDetails(details); err != nil {
			return model.Purchase{}, fmt.Errorf("%w: purchase_item_id %s: %v", ErrInvalidInput, item.ID, err)
		}
		fulfillmentID, err := randomID("fulfillment")
		if err != nil {
			return model.Purchase{}, fmt.Errorf("generate fulfillment id: %w", err)
		}
		prepared = append(prepared, model.MerchantFulfillment{
			ID: fulfillmentID, PurchaseItemID: item.ID,
			Status: model.MerchantFulfillmentStatusPending, IdempotencyKey: fulfillmentID,
			DetailsSnapshot: cloneJSON(details), CreatedAt: now, UpdatedAt: now,
		})
	}
	if err := s.repository.PrepareFulfillments(ctx, prepared); err != nil {
		return model.Purchase{}, err
	}

	fulfillments, err := s.repository.ListFulfillments(ctx, purchase.ID)
	if err != nil {
		return model.Purchase{}, err
	}

	// Stripe導入時は、Merchant注文確定より前にPaymentIntentの認証・確定を接続する。
	for _, value := range fulfillments {
		if value.Status == model.MerchantFulfillmentStatusConfirmed || value.Status == model.MerchantFulfillmentStatusRejected {
			continue
		}
		item, ok := items[value.PurchaseItemID]
		if !ok {
			return model.Purchase{}, fmt.Errorf("purchase item %s is missing", value.PurchaseItemID)
		}
		if err := s.process(ctx, purchase, item, value); err != nil {
			return model.Purchase{}, err
		}
	}

	fulfillments, err = s.repository.ListFulfillments(ctx, purchase.ID)
	if err != nil {
		return model.Purchase{}, err
	}
	target := aggregateStatus(fulfillments, len(purchase.Items))
	if target != model.PurchaseStatusCreated {
		if err := s.repository.UpdatePurchaseStatus(ctx, purchase.ID, model.PurchaseStatusCreated, target); err != nil {
			return model.Purchase{}, err
		}
	}
	return s.repository.FindPurchase(ctx, purchase.ID)
}

func (s *Service) process(ctx context.Context, purchase model.Purchase, item model.PurchaseItem, value model.MerchantFulfillment) error {
	capability, err := s.merchants.FindByID(ctx, item.CapabilityID)
	if errors.Is(err, merchant.ErrNotFound) {
		return s.storeResult(ctx, value, MerchantResult{Status: model.MerchantFulfillmentStatusRejected, FailureCode: "capability_unavailable"})
	}
	if err != nil {
		return fmt.Errorf("find merchant capability: %w", err)
	}
	if capability.Merchant.ID != item.MerchantID || capability.Domain != item.Domain {
		return s.storeResult(ctx, value, MerchantResult{Status: model.MerchantFulfillmentStatusRejected, FailureCode: "capability_changed"})
	}
	fulfiller, ok := s.fulfillers.Fulfiller(item.Domain)
	if !ok {
		return s.storeResult(ctx, value, MerchantResult{Status: model.MerchantFulfillmentStatusRejected, FailureCode: "domain_fulfillment_unavailable"})
	}
	request := MerchantRequest{
		PurchaseID: purchase.ID, PurchaseItemID: item.ID, BuyerRef: purchase.BuyerRef,
		MerchantOfferRef: item.MerchantOfferRef, IdempotencyKey: value.IdempotencyKey,
		Details: cloneJSON(value.DetailsSnapshot),
	}
	var result MerchantResult
	if value.Status == model.MerchantFulfillmentStatusUnknown {
		result, err = fulfiller.Resolve(ctx, capability, value.IdempotencyKey)
		if errors.Is(err, ErrMerchantRecordNotFound) {
			result, err = fulfiller.Submit(ctx, capability, request)
		}
	} else {
		result, err = fulfiller.Submit(ctx, capability, request)
	}
	if err != nil {
		result = MerchantResult{Status: model.MerchantFulfillmentStatusUnknown, FailureCode: "merchant_response_unknown"}
	}
	return s.storeResult(ctx, value, result)
}

func (s *Service) storeResult(ctx context.Context, value model.MerchantFulfillment, result MerchantResult) error {
	if !validMerchantStatus(result.Status) {
		result = MerchantResult{Status: model.MerchantFulfillmentStatusUnknown, FailureCode: "invalid_merchant_response"}
	}
	expectedStatus := value.Status
	value.Status = result.Status
	value.MerchantOrderRef = result.MerchantOrderRef
	value.FailureCode = result.FailureCode
	value.ResponseSnapshot = cloneJSON(result.Snapshot)
	value.UpdatedAt = time.Now().UTC()
	return s.repository.UpdateFulfillment(ctx, value, expectedStatus)
}

func aggregateStatus(values []model.MerchantFulfillment, itemCount int) model.PurchaseStatus {
	if itemCount == 0 || len(values) != itemCount {
		return model.PurchaseStatusCreated
	}
	confirmed, rejected := 0, 0
	for _, value := range values {
		switch value.Status {
		case model.MerchantFulfillmentStatusConfirmed:
			confirmed++
		case model.MerchantFulfillmentStatusRejected:
			rejected++
		}
	}
	if confirmed == itemCount {
		return model.PurchaseStatusConfirmed
	}
	if rejected == itemCount {
		return model.PurchaseStatusCancelled
	}
	return model.PurchaseStatusCreated
}

func validMerchantStatus(status model.MerchantFulfillmentStatus) bool {
	switch status {
	case model.MerchantFulfillmentStatusPending, model.MerchantFulfillmentStatusConfirmed,
		model.MerchantFulfillmentStatusRejected, model.MerchantFulfillmentStatusUnknown:
		return true
	default:
		return false
	}
}

func randomID(prefix string) (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(value[:]), nil
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	if value == nil {
		return nil
	}
	return append(json.RawMessage(nil), value...)
}
