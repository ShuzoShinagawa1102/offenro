package cart

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

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
		ID: cartID, AgentID: input.AgentID, BuyerRef: input.BuyerRef,
		Status: model.CartStatusActive, Items: []model.CartItem{},
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.carts.CreateCart(ctx, created); err != nil {
		return model.Cart{}, err
	}
	return created, nil
}

func (s *Service) Get(ctx context.Context, cartID string) (model.Cart, error) {
	if strings.TrimSpace(cartID) == "" {
		return model.Cart{}, fmt.Errorf("%w: cart_id is required", ErrInvalidInput)
	}
	return s.carts.GetCart(ctx, cartID)
}

func (s *Service) Update(ctx context.Context, cartID string, input UpdateInput) (model.Cart, error) {
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
	if err := s.carts.UpdateCart(ctx, current, expectedUpdatedAt); err != nil {
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
	return s.carts.UpdateCart(ctx, current, expectedUpdatedAt)
}
