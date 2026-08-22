package cart

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
)

func (s *Service) Checkout(ctx context.Context, cartID string) (model.Purchase, error) {
	current, err := s.Get(ctx, cartID)
	if err != nil {
		return model.Purchase{}, err
	}
	if current.Status != model.CartStatusActive {
		return model.Purchase{}, fmt.Errorf("%w: only an active cart can be checked out", ErrConflict)
	}
	if s.offers == nil || s.merchants == nil || s.verifiers == nil {
		return model.Purchase{}, fmt.Errorf("%w: checkout dependencies are unavailable", ErrRevalidation)
	}

	activeItems := make([]model.CartItem, 0, len(current.Items))
	for _, item := range current.Items {
		if item.Status == model.CartItemStatusActive {
			activeItems = append(activeItems, item)
		}
	}
	if len(activeItems) == 0 {
		return model.Purchase{}, fmt.Errorf("%w: cart has no active items", ErrInvalidInput)
	}

	verifiedItems := make([]model.CartItem, 0, len(activeItems))
	for _, item := range activeItems {
		verified, verifyErr := s.revalidateItem(ctx, item)
		if verifyErr != nil {
			return model.Purchase{}, verifyErr
		}
		verifiedItems = append(verifiedItems, verified)
	}

	purchaseID, err := newID("purchase")
	if err != nil {
		return model.Purchase{}, fmt.Errorf("generate purchase id: %w", err)
	}
	purchasedAt := time.Now().UTC()
	purchase := model.Purchase{
		ID: purchaseID, CartID: current.ID, AgentID: current.AgentID, BuyerRef: current.BuyerRef,
		Status: model.PurchaseStatusCreated, Items: make([]model.PurchaseItem, 0, len(verifiedItems)),
		Currency: verifiedItems[0].Currency, PurchasedAt: purchasedAt,
	}
	for _, item := range verifiedItems {
		if item.Currency != purchase.Currency {
			return model.Purchase{}, fmt.Errorf("%w: all cart items must use the same currency", ErrInvalidInput)
		}
		if item.Amount > math.MaxInt64-purchase.TotalAmount {
			return model.Purchase{}, fmt.Errorf("%w: total amount overflow", ErrInvalidInput)
		}
		purchaseItemID, idErr := newID("purchase_item")
		if idErr != nil {
			return model.Purchase{}, fmt.Errorf("generate purchase item id: %w", idErr)
		}
		purchase.TotalAmount += item.Amount
		purchase.Items = append(purchase.Items, model.PurchaseItem{
			ID: purchaseItemID, PurchaseID: purchase.ID, CapabilityID: item.CapabilityID,
			MerchantID: item.MerchantID, Domain: item.Domain, OfferID: item.OfferID,
			MerchantOfferRef: item.MerchantOfferRef, OfferSnapshot: cloneJSON(item.OfferSnapshot),
			Amount: item.Amount, Currency: item.Currency,
		})
	}
	if err := s.carts.Checkout(ctx, purchase, current.UpdatedAt, purchasedAt); err != nil {
		return model.Purchase{}, err
	}
	return purchase, nil
}

func (s *Service) revalidateItem(ctx context.Context, item model.CartItem) (model.CartItem, error) {
	reference, err := s.offers.Resolve(item.OfferID)
	if err != nil {
		if errors.Is(err, offer.ErrExpiredReference) {
			return model.CartItem{}, fmt.Errorf("%w: offer has expired", ErrConflict)
		}
		return model.CartItem{}, fmt.Errorf("%w: stored offer reference is invalid", ErrConflict)
	}
	if reference.CapabilityID != item.CapabilityID || reference.MerchantID != item.MerchantID ||
		reference.Domain != item.Domain || reference.MerchantOfferRef != item.MerchantOfferRef {
		return model.CartItem{}, fmt.Errorf("%w: stored offer does not match its reference", ErrConflict)
	}
	capability, err := s.merchants.FindByID(ctx, item.CapabilityID)
	if err != nil {
		return model.CartItem{}, fmt.Errorf("%w: merchant capability is no longer active", ErrConflict)
	}
	if capability.Merchant.ID != item.MerchantID || capability.Domain != item.Domain {
		return model.CartItem{}, fmt.Errorf("%w: merchant capability changed", ErrConflict)
	}
	verifier, ok := s.verifiers.OfferVerifier(item.Domain)
	if !ok {
		return model.CartItem{}, fmt.Errorf("%w: verifier is unavailable for domain %s", ErrRevalidation, item.Domain)
	}
	verified, err := verifier.Revalidate(ctx, capability, item.OfferID, item.MerchantOfferRef)
	if err != nil {
		if errors.Is(err, offer.ErrUnavailable) || errors.Is(err, offer.ErrExpiredReference) {
			return model.CartItem{}, fmt.Errorf("%w: merchant offer is no longer available", ErrConflict)
		}
		return model.CartItem{}, fmt.Errorf("%w: %v", ErrRevalidation, err)
	}
	if !validSnapshot(verified.Snapshot) || verified.Amount < 0 || !validCurrency(verified.Currency) {
		return model.CartItem{}, fmt.Errorf("%w: merchant returned an invalid offer", ErrRevalidation)
	}
	if verified.Amount != item.Amount || normalizeCurrency(verified.Currency) != item.Currency {
		return model.CartItem{}, fmt.Errorf("%w: merchant offer price changed", ErrConflict)
	}
	if verified.ExpiresAt != nil && !verified.ExpiresAt.After(time.Now()) {
		return model.CartItem{}, fmt.Errorf("%w: merchant offer has expired", ErrConflict)
	}
	item.OfferSnapshot = cloneJSON(verified.Snapshot)
	item.Amount = verified.Amount
	item.Currency = normalizeCurrency(verified.Currency)
	item.OfferExpiresAt = verified.ExpiresAt
	return item, nil
}

func (s *Service) GetPurchase(ctx context.Context, purchaseID string) (model.Purchase, error) {
	if strings.TrimSpace(purchaseID) == "" {
		return model.Purchase{}, fmt.Errorf("%w: purchase_id is required", ErrInvalidInput)
	}
	return s.purchases.GetPurchase(ctx, purchaseID)
}
