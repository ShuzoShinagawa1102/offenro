package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/cart"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	postgresdb "github.com/ShuzoShinagawa1102/offenro/internal/platform/postgres/generated"
	"github.com/jackc/pgx/v5"
)

var _ cart.Repository = (*Store)(nil)

func (s *Store) CreateCart(ctx context.Context, value model.Cart) error {
	rowsAffected, err := s.queries.CreateCart(ctx, postgresdb.CreateCartParams{
		CartID: value.ID, BuyerRef: value.BuyerRef, Status: string(value.Status),
		CreatedAt: timestamp(value.CreatedAt), UpdatedAt: timestamp(value.UpdatedAt), AgentID: value.AgentID,
	})
	if err != nil {
		return fmt.Errorf("create cart: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: active agent not found", cart.ErrInvalidInput)
	}
	return nil
}

func (s *Store) GetCart(ctx context.Context, cartID string) (model.Cart, error) {
	row, err := s.queries.GetCart(ctx, cartID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Cart{}, fmt.Errorf("%w: cart", cart.ErrNotFound)
	}
	if err != nil {
		return model.Cart{}, fmt.Errorf("get cart: %w", err)
	}
	itemRows, err := s.queries.ListCartItems(ctx, cartID)
	if err != nil {
		return model.Cart{}, fmt.Errorf("list cart items: %w", err)
	}

	value := model.Cart{
		ID: row.CartID, AgentID: row.AgentID, BuyerRef: row.BuyerRef,
		Status: model.CartStatus(row.Status), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
		Items: make([]model.CartItem, 0, len(itemRows)),
	}
	for _, item := range itemRows {
		value.Items = append(value.Items, model.CartItem{
			ID: item.CartItemID, CartID: item.CartID, CapabilityID: item.CapabilityID,
			MerchantID: model.MerchantID(item.MerchantID), Domain: model.Domain(item.DomainID),
			OfferID: item.OfferID, MerchantOfferRef: item.MerchantOfferRef,
			OfferSnapshot: item.OfferSnapshot, Amount: item.Amount,
			Currency: item.Currency, OfferExpiresAt: timestampPointer(item.OfferExpiresAt),
			Status: model.CartItemStatus(item.Status),
		})
	}
	return value, nil
}

func (s *Store) UpdateCart(ctx context.Context, value model.Cart, expectedUpdatedAt time.Time) error {
	rowsAffected, err := s.queries.UpdateCart(ctx, postgresdb.UpdateCartParams{
		BuyerRef: value.BuyerRef, Status: string(value.Status), UpdatedAt: timestamp(value.UpdatedAt),
		CartID: value.ID, ExpectedUpdatedAt: timestamp(expectedUpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("update cart: %w", err)
	}
	return requireSingleActiveCart(rowsAffected)
}

func (s *Store) AddCartItem(
	ctx context.Context,
	item model.CartItem,
	expectedUpdatedAt time.Time,
	updatedAt time.Time,
) (model.CartItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return model.CartItem{}, fmt.Errorf("begin add cart item transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := s.queries.WithTx(tx)

	if err := touchActiveCart(ctx, queries, item.CartID, expectedUpdatedAt, updatedAt); err != nil {
		return model.CartItem{}, err
	}
	item.CapabilityID, err = queries.CreateCartItem(ctx, postgresdb.CreateCartItemParams{
		CartItemID: item.ID, CartID: item.CartID, OfferID: item.OfferID,
		MerchantOfferRef: item.MerchantOfferRef,
		OfferSnapshot:    item.OfferSnapshot, OfferExpiresAt: optionalTimestamp(item.OfferExpiresAt),
		Status: string(item.Status), Amount: item.Amount, Currency: item.Currency,
		CapabilityID: item.CapabilityID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.CartItem{}, fmt.Errorf("%w: active merchant capability not found", cart.ErrInvalidInput)
	}
	if err != nil {
		return model.CartItem{}, fmt.Errorf("create cart item: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.CartItem{}, fmt.Errorf("commit cart item: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateCartItem(
	ctx context.Context,
	item model.CartItem,
	expectedUpdatedAt time.Time,
	updatedAt time.Time,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update cart item transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := s.queries.WithTx(tx)

	if err := touchActiveCart(ctx, queries, item.CartID, expectedUpdatedAt, updatedAt); err != nil {
		return err
	}
	rowsAffected, err := queries.UpdateCartItem(ctx, postgresdb.UpdateCartItemParams{
		OfferSnapshot: item.OfferSnapshot, OfferExpiresAt: optionalTimestamp(item.OfferExpiresAt),
		Status: string(item.Status), Amount: item.Amount, Currency: item.Currency,
		CartItemID: item.ID, CartID: item.CartID,
	})
	if err != nil {
		return fmt.Errorf("update cart item: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: cart item changed concurrently", cart.ErrConflict)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit cart item update: %w", err)
	}
	return nil
}

func (s *Store) Checkout(
	ctx context.Context,
	purchase model.Purchase,
	expectedUpdatedAt time.Time,
	updatedAt time.Time,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin checkout transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := s.queries.WithTx(tx)

	rowsAffected, err := queries.MarkCartCheckedOut(ctx, postgresdb.MarkCartCheckedOutParams{
		UpdatedAt: timestamp(updatedAt), CartID: purchase.CartID, ExpectedUpdatedAt: timestamp(expectedUpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("mark cart checked out: %w", err)
	}
	if err := requireSingleActiveCart(rowsAffected); err != nil {
		return err
	}
	if err := queries.CreatePurchase(ctx, postgresdb.CreatePurchaseParams{
		PurchaseID: purchase.ID, CartID: purchase.CartID, AgentID: purchase.AgentID,
		BuyerRef: purchase.BuyerRef, Status: string(purchase.Status), TotalAmount: purchase.TotalAmount,
		Currency: purchase.Currency, PurchasedAt: timestamp(purchase.PurchasedAt),
	}); err != nil {
		return fmt.Errorf("create purchase: %w", err)
	}
	for _, item := range purchase.Items {
		if err := queries.CreatePurchaseItem(ctx, postgresdb.CreatePurchaseItemParams{
			PurchaseItemID: item.ID, PurchaseID: purchase.ID, CapabilityID: item.CapabilityID,
			OfferID: item.OfferID, MerchantOfferRef: item.MerchantOfferRef,
			OfferSnapshot: item.OfferSnapshot, Amount: item.Amount, Currency: item.Currency,
		}); err != nil {
			return fmt.Errorf("create purchase item: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit checkout: %w", err)
	}
	return nil
}

func (s *Store) GetPurchase(ctx context.Context, purchaseID string) (model.Purchase, error) {
	row, err := s.queries.GetPurchase(ctx, purchaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Purchase{}, fmt.Errorf("%w: purchase", cart.ErrNotFound)
	}
	if err != nil {
		return model.Purchase{}, fmt.Errorf("get purchase: %w", err)
	}
	itemRows, err := s.queries.ListPurchaseItems(ctx, purchaseID)
	if err != nil {
		return model.Purchase{}, fmt.Errorf("list purchase items: %w", err)
	}

	value := model.Purchase{
		ID: row.PurchaseID, CartID: row.CartID, AgentID: row.AgentID, BuyerRef: row.BuyerRef,
		Status: model.PurchaseStatus(row.Status), TotalAmount: row.TotalAmount,
		Currency: row.Currency, PurchasedAt: row.PurchasedAt.Time,
		Items: make([]model.PurchaseItem, 0, len(itemRows)),
	}
	for _, item := range itemRows {
		value.Items = append(value.Items, model.PurchaseItem{
			ID: item.PurchaseItemID, PurchaseID: item.PurchaseID, CapabilityID: item.CapabilityID,
			MerchantID: model.MerchantID(item.MerchantID), Domain: model.Domain(item.DomainID),
			OfferID: item.OfferID, MerchantOfferRef: item.MerchantOfferRef,
			OfferSnapshot: item.OfferSnapshot, Amount: item.Amount, Currency: item.Currency,
		})
	}
	return value, nil
}

func touchActiveCart(
	ctx context.Context,
	queries *postgresdb.Queries,
	cartID string,
	expectedUpdatedAt time.Time,
	updatedAt time.Time,
) error {
	rowsAffected, err := queries.TouchActiveCart(ctx, postgresdb.TouchActiveCartParams{
		UpdatedAt: timestamp(updatedAt), CartID: cartID, ExpectedUpdatedAt: timestamp(expectedUpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("touch active cart: %w", err)
	}
	return requireSingleActiveCart(rowsAffected)
}

func requireSingleActiveCart(rowsAffected int64) error {
	if rowsAffected != 1 {
		return fmt.Errorf("%w: cart is no longer active or changed concurrently", cart.ErrConflict)
	}
	return nil
}
