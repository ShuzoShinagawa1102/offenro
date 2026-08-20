package cart

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type Repository interface {
	CreateCart(ctx context.Context, cart model.Cart) error
	GetCart(ctx context.Context, cartID string) (model.Cart, error)
	UpdateCart(
		ctx context.Context,
		cart model.Cart,
		expectedUpdatedAt time.Time,
	) error
	AddCartItem(
		ctx context.Context,
		item model.CartItem,
		expectedUpdatedAt time.Time,
		updatedAt time.Time,
	) (model.CartItem, error)
	UpdateCartItem(
		ctx context.Context,
		item model.CartItem,
		expectedUpdatedAt time.Time,
		updatedAt time.Time,
	) error
	Checkout(
		ctx context.Context,
		purchase model.Purchase,
		expectedUpdatedAt time.Time,
		updatedAt time.Time,
	) error
	GetPurchase(ctx context.Context, purchaseID string) (model.Purchase, error)
}

type CreateInput struct {
	AgentID  string
	BuyerRef string
}

type UpdateInput struct {
	BuyerRef string
}

type AddItemInput struct {
	MerchantID     model.MerchantID
	Domain         model.Domain
	OfferID        string
	OfferSnapshot  json.RawMessage
	Amount         int64
	Currency       string
	OfferExpiresAt *time.Time
}

type UpdateItemInput struct {
	OfferSnapshot  *json.RawMessage
	Amount         *int64
	Currency       *string
	OfferExpiresAt *time.Time
}

type UseCases interface {
	Create(ctx context.Context, input CreateInput) (model.Cart, error)
	Get(ctx context.Context, cartID string) (model.Cart, error)
	Update(ctx context.Context, cartID string, input UpdateInput) (model.Cart, error)
	Delete(ctx context.Context, cartID string) error
	AddItem(ctx context.Context, cartID string, input AddItemInput) (model.CartItem, error)
	UpdateItem(
		ctx context.Context,
		cartID string,
		itemID string,
		input UpdateItemInput,
	) (model.CartItem, error)
	DeleteItem(ctx context.Context, cartID string, itemID string) error
	Checkout(ctx context.Context, cartID string) (model.Purchase, error)
	GetPurchase(ctx context.Context, purchaseID string) (model.Purchase, error)
}
