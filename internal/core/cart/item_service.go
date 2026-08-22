package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
)

func (s *Service) AddItem(ctx context.Context, cartID string, input AddItemInput) (model.CartItem, error) {
	if err := validateNewItem(input); err != nil {
		return model.CartItem{}, err
	}
	if s.offers == nil {
		return model.CartItem{}, fmt.Errorf("offer resolver is unavailable")
	}
	reference, err := s.offers.Resolve(input.OfferID)
	if err != nil {
		if errors.Is(err, offer.ErrExpiredReference) {
			return model.CartItem{}, fmt.Errorf("%w: offer_id has expired", ErrConflict)
		}
		return model.CartItem{}, fmt.Errorf("%w: invalid offer_id", ErrInvalidInput)
	}
	current, err := s.Get(ctx, cartID)
	if err != nil {
		return model.CartItem{}, err
	}
	if current.Status != model.CartStatusActive {
		return model.CartItem{}, fmt.Errorf("%w: items can only be added to an active cart", ErrConflict)
	}

	itemID, err := newID("cart_item")
	if err != nil {
		return model.CartItem{}, fmt.Errorf("generate cart item id: %w", err)
	}
	expiresAt := reference.ExpiresAt.UTC()
	item := model.CartItem{
		ID: itemID, CartID: current.ID, CapabilityID: reference.CapabilityID,
		MerchantID: reference.MerchantID, Domain: reference.Domain,
		OfferID: input.OfferID, MerchantOfferRef: reference.MerchantOfferRef,
		OfferSnapshot: cloneJSON(input.OfferSnapshot), Amount: input.Amount,
		Currency: normalizeCurrency(input.Currency), OfferExpiresAt: &expiresAt,
		Status: model.CartItemStatusActive,
	}
	return s.carts.AddCartItem(ctx, item, current.UpdatedAt, time.Now().UTC())
}

func (s *Service) UpdateItem(ctx context.Context, cartID string, itemID string, input UpdateItemInput) (model.CartItem, error) {
	if input.OfferSnapshot == nil && input.Amount == nil && input.Currency == nil && input.OfferExpiresAt == nil {
		return model.CartItem{}, fmt.Errorf("%w: at least one item field is required", ErrInvalidInput)
	}
	if input.OfferSnapshot != nil && !validSnapshot(*input.OfferSnapshot) {
		return model.CartItem{}, fmt.Errorf("%w: offer_snapshot must be a JSON object", ErrInvalidInput)
	}
	if input.Amount != nil && *input.Amount < 0 {
		return model.CartItem{}, fmt.Errorf("%w: amount must not be negative", ErrInvalidInput)
	}
	if input.Currency != nil && !validCurrency(*input.Currency) {
		return model.CartItem{}, fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidInput)
	}
	if input.OfferExpiresAt != nil && !input.OfferExpiresAt.After(time.Now()) {
		return model.CartItem{}, fmt.Errorf("%w: offer_expires_at must be in the future", ErrInvalidInput)
	}

	current, err := s.Get(ctx, cartID)
	if err != nil {
		return model.CartItem{}, err
	}
	if current.Status != model.CartStatusActive {
		return model.CartItem{}, fmt.Errorf("%w: items can only be updated in an active cart", ErrConflict)
	}
	item, found := findItem(current.Items, itemID)
	if !found {
		return model.CartItem{}, fmt.Errorf("%w: cart item", ErrNotFound)
	}
	if item.Status != model.CartItemStatusActive {
		return model.CartItem{}, fmt.Errorf("%w: only an active cart item can be updated", ErrConflict)
	}
	if input.OfferSnapshot != nil {
		item.OfferSnapshot = cloneJSON(*input.OfferSnapshot)
	}
	if input.Amount != nil {
		item.Amount = *input.Amount
	}
	if input.Currency != nil {
		item.Currency = normalizeCurrency(*input.Currency)
	}
	if input.OfferExpiresAt != nil {
		expiresAt := input.OfferExpiresAt.UTC()
		item.OfferExpiresAt = &expiresAt
	}
	updatedAt := time.Now().UTC()
	if err := s.carts.UpdateCartItem(ctx, item, current.UpdatedAt, updatedAt); err != nil {
		return model.CartItem{}, err
	}
	return item, nil
}

func (s *Service) DeleteItem(ctx context.Context, cartID string, itemID string) error {
	current, err := s.Get(ctx, cartID)
	if err != nil {
		return err
	}
	if current.Status != model.CartStatusActive {
		return fmt.Errorf("%w: items can only be removed from an active cart", ErrConflict)
	}
	item, found := findItem(current.Items, itemID)
	if !found {
		return fmt.Errorf("%w: cart item", ErrNotFound)
	}
	if item.Status != model.CartItemStatusActive {
		return fmt.Errorf("%w: only an active cart item can be removed", ErrConflict)
	}
	item.Status = model.CartItemStatusRemoved
	return s.carts.UpdateCartItem(ctx, item, current.UpdatedAt, time.Now().UTC())
}

func findItem(items []model.CartItem, itemID string) (model.CartItem, bool) {
	for _, item := range items {
		if item.ID == itemID {
			return item, true
		}
	}
	return model.CartItem{}, false
}
