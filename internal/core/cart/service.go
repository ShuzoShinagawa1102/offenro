package cart

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type Service struct {
	repository Repository
}

var _ UseCases = (*Service)(nil)

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (model.Cart, error) {
	input.AgentID = strings.TrimSpace(input.AgentID)
	input.BuyerRef = strings.TrimSpace(input.BuyerRef)
	if input.AgentID == "" || input.BuyerRef == "" {
		return model.Cart{}, fmt.Errorf("%w: agent_id and buyer_ref are required", ErrInvalidInput)
	}

	cartID, err := newID("cart")
	if err != nil {
		return model.Cart{}, fmt.Errorf("generate cart id: %w", err)
	}
	now := time.Now().UTC()
	created := model.Cart{
		ID:        cartID,
		AgentID:   input.AgentID,
		BuyerRef:  input.BuyerRef,
		Status:    model.CartStatusActive,
		Items:     []model.CartItem{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repository.CreateCart(ctx, created); err != nil {
		return model.Cart{}, err
	}

	return created, nil
}

func (s *Service) Get(ctx context.Context, cartID string) (model.Cart, error) {
	if strings.TrimSpace(cartID) == "" {
		return model.Cart{}, fmt.Errorf("%w: cart_id is required", ErrInvalidInput)
	}
	return s.repository.GetCart(ctx, cartID)
}

func (s *Service) Update(
	ctx context.Context,
	cartID string,
	input UpdateInput,
) (model.Cart, error) {
	buyerRef := strings.TrimSpace(input.BuyerRef)
	if buyerRef == "" {
		return model.Cart{}, fmt.Errorf("%w: buyer_ref is required", ErrInvalidInput)
	}

	current, err := s.Get(ctx, cartID)
	if err != nil {
		return model.Cart{}, err
	}
	if current.Status != model.CartStatusActive {
		return model.Cart{}, fmt.Errorf("%w: only an active cart can be updated", ErrConflict)
	}

	expectedUpdatedAt := current.UpdatedAt
	current.BuyerRef = buyerRef
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.UpdateCart(ctx, current, expectedUpdatedAt); err != nil {
		return model.Cart{}, err
	}

	return current, nil
}

func (s *Service) Delete(ctx context.Context, cartID string) error {
	current, err := s.Get(ctx, cartID)
	if err != nil {
		return err
	}
	if current.Status != model.CartStatusActive {
		return fmt.Errorf("%w: only an active cart can be abandoned", ErrConflict)
	}

	expectedUpdatedAt := current.UpdatedAt
	current.Status = model.CartStatusAbandoned
	current.UpdatedAt = time.Now().UTC()
	return s.repository.UpdateCart(ctx, current, expectedUpdatedAt)
}

func (s *Service) AddItem(
	ctx context.Context,
	cartID string,
	input AddItemInput,
) (model.CartItem, error) {
	if err := validateNewItem(input); err != nil {
		return model.CartItem{}, err
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
	updatedAt := time.Now().UTC()
	item := model.CartItem{
		ID:             itemID,
		CartID:         current.ID,
		MerchantID:     input.MerchantID,
		Domain:         input.Domain,
		OfferID:        strings.TrimSpace(input.OfferID),
		OfferSnapshot:  cloneJSON(input.OfferSnapshot),
		Amount:         input.Amount,
		Currency:       normalizeCurrency(input.Currency),
		OfferExpiresAt: input.OfferExpiresAt,
		Status:         model.CartItemStatusActive,
	}

	return s.repository.AddCartItem(ctx, item, current.UpdatedAt, updatedAt)
}

func (s *Service) UpdateItem(
	ctx context.Context,
	cartID string,
	itemID string,
	input UpdateItemInput,
) (model.CartItem, error) {
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
	if err := s.repository.UpdateCartItem(ctx, item, current.UpdatedAt, updatedAt); err != nil {
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
	return s.repository.UpdateCartItem(
		ctx,
		item,
		current.UpdatedAt,
		time.Now().UTC(),
	)
}

func (s *Service) Checkout(ctx context.Context, cartID string) (model.Purchase, error) {
	current, err := s.Get(ctx, cartID)
	if err != nil {
		return model.Purchase{}, err
	}
	if current.Status != model.CartStatusActive {
		return model.Purchase{}, fmt.Errorf("%w: only an active cart can be checked out", ErrConflict)
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

	purchaseID, err := newID("purchase")
	if err != nil {
		return model.Purchase{}, fmt.Errorf("generate purchase id: %w", err)
	}
	purchasedAt := time.Now().UTC()
	purchase := model.Purchase{
		ID:          purchaseID,
		CartID:      current.ID,
		AgentID:     current.AgentID,
		BuyerRef:    current.BuyerRef,
		Status:      model.PurchaseStatusCreated,
		Items:       make([]model.PurchaseItem, 0, len(activeItems)),
		Currency:    activeItems[0].Currency,
		PurchasedAt: purchasedAt,
	}
	for _, item := range activeItems {
		if item.OfferExpiresAt != nil && !item.OfferExpiresAt.After(purchasedAt) {
			return model.Purchase{}, fmt.Errorf("%w: cart contains an expired offer", ErrConflict)
		}
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
			ID:            purchaseItemID,
			PurchaseID:    purchase.ID,
			CapabilityID:  item.CapabilityID,
			MerchantID:    item.MerchantID,
			Domain:        item.Domain,
			OfferID:       item.OfferID,
			OfferSnapshot: cloneJSON(item.OfferSnapshot),
			Amount:        item.Amount,
			Currency:      item.Currency,
		})
	}

	if err := s.repository.Checkout(
		ctx,
		purchase,
		current.UpdatedAt,
		purchasedAt,
	); err != nil {
		return model.Purchase{}, err
	}
	return purchase, nil
}

func (s *Service) GetPurchase(
	ctx context.Context,
	purchaseID string,
) (model.Purchase, error) {
	if strings.TrimSpace(purchaseID) == "" {
		return model.Purchase{}, fmt.Errorf("%w: purchase_id is required", ErrInvalidInput)
	}
	return s.repository.GetPurchase(ctx, purchaseID)
}

func validateNewItem(input AddItemInput) error {
	if strings.TrimSpace(string(input.MerchantID)) == "" ||
		strings.TrimSpace(input.Domain.String()) == "" ||
		strings.TrimSpace(input.OfferID) == "" {
		return fmt.Errorf("%w: merchant_id, domain and offer_id are required", ErrInvalidInput)
	}
	if !validSnapshot(input.OfferSnapshot) {
		return fmt.Errorf("%w: offer_snapshot must be a JSON object", ErrInvalidInput)
	}
	if input.Amount < 0 {
		return fmt.Errorf("%w: amount must not be negative", ErrInvalidInput)
	}
	if !validCurrency(input.Currency) {
		return fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidInput)
	}
	if input.OfferExpiresAt != nil && !input.OfferExpiresAt.After(time.Now()) {
		return fmt.Errorf("%w: offer_expires_at must be in the future", ErrInvalidInput)
	}
	return nil
}

func findItem(items []model.CartItem, itemID string) (model.CartItem, bool) {
	for _, item := range items {
		if item.ID == itemID {
			return item, true
		}
	}
	return model.CartItem{}, false
}

func validSnapshot(snapshot json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(snapshot))
	return len(trimmed) > 1 && trimmed[0] == '{' && json.Valid(snapshot)
}

func validCurrency(currency string) bool {
	normalized := normalizeCurrency(currency)
	if len(normalized) != 3 {
		return false
	}
	for _, character := range normalized {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), value...)
}

func newID(prefix string) (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(value[:]), nil
}
