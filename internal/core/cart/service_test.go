package cart

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type testRepository struct {
	cart     model.Cart
	purchase model.Purchase
}

func (r *testRepository) CreateCart(_ context.Context, value model.Cart) error {
	r.cart = value
	return nil
}

func (r *testRepository) GetCart(_ context.Context, cartID string) (model.Cart, error) {
	if r.cart.ID != cartID {
		return model.Cart{}, ErrNotFound
	}
	value := r.cart
	value.Items = append([]model.CartItem(nil), r.cart.Items...)
	return value, nil
}

func (r *testRepository) UpdateCart(
	_ context.Context,
	value model.Cart,
	expectedUpdatedAt time.Time,
) error {
	if r.cart.UpdatedAt != expectedUpdatedAt {
		return ErrConflict
	}
	r.cart = value
	return nil
}

func (r *testRepository) AddCartItem(
	_ context.Context,
	item model.CartItem,
	expectedUpdatedAt time.Time,
	updatedAt time.Time,
) (model.CartItem, error) {
	if r.cart.UpdatedAt != expectedUpdatedAt {
		return model.CartItem{}, ErrConflict
	}
	item.CapabilityID = "capability-1"
	r.cart.Items = append(r.cart.Items, item)
	r.cart.UpdatedAt = updatedAt
	return item, nil
}

func (r *testRepository) UpdateCartItem(
	_ context.Context,
	item model.CartItem,
	expectedUpdatedAt time.Time,
	updatedAt time.Time,
) error {
	if r.cart.UpdatedAt != expectedUpdatedAt {
		return ErrConflict
	}
	for index := range r.cart.Items {
		if r.cart.Items[index].ID == item.ID {
			r.cart.Items[index] = item
			r.cart.UpdatedAt = updatedAt
			return nil
		}
	}
	return ErrNotFound
}

func (r *testRepository) Checkout(
	_ context.Context,
	purchase model.Purchase,
	expectedUpdatedAt time.Time,
	updatedAt time.Time,
) error {
	if r.cart.UpdatedAt != expectedUpdatedAt || r.cart.Status != model.CartStatusActive {
		return ErrConflict
	}
	r.cart.Status = model.CartStatusCheckedOut
	r.cart.UpdatedAt = updatedAt
	r.purchase = purchase
	return nil
}

func (r *testRepository) GetPurchase(
	_ context.Context,
	purchaseID string,
) (model.Purchase, error) {
	if r.purchase.ID != purchaseID {
		return model.Purchase{}, ErrNotFound
	}
	return r.purchase, nil
}

func TestServiceCheckoutCreatesPurchaseFromActiveItems(t *testing.T) {
	t.Parallel()

	repository := &testRepository{}
	service := NewService(repository)
	ctx := context.Background()

	created, err := service.Create(ctx, CreateInput{
		AgentID:  "agent-1",
		BuyerRef: "buyer-1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	item, err := service.AddItem(ctx, created.ID, AddItemInput{
		MerchantID:    "merchant-1",
		Domain:        "travel.hotel",
		OfferID:       "offer-1",
		OfferSnapshot: json.RawMessage(`{"offer_id":"offer-1"}`),
		Amount:        12000,
		Currency:      "jpy",
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if item.Currency != "JPY" {
		t.Fatalf("item currency = %q, want JPY", item.Currency)
	}

	purchase, err := service.Checkout(ctx, created.ID)
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}
	if purchase.Status != model.PurchaseStatusCreated {
		t.Fatalf("purchase status = %q, want %q", purchase.Status, model.PurchaseStatusCreated)
	}
	if purchase.TotalAmount != 12000 || purchase.Currency != "JPY" {
		t.Fatalf(
			"purchase money = %d %s, want 12000 JPY",
			purchase.TotalAmount,
			purchase.Currency,
		)
	}
	if len(purchase.Items) != 1 || purchase.Items[0].CapabilityID != "capability-1" {
		t.Fatalf("purchase items = %#v", purchase.Items)
	}

	_, err = service.Checkout(ctx, created.ID)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("second Checkout() error = %v, want conflict", err)
	}
}

func TestServiceCheckoutRejectsMixedCurrencies(t *testing.T) {
	t.Parallel()

	repository := &testRepository{}
	service := NewService(repository)
	ctx := context.Background()
	created, err := service.Create(ctx, CreateInput{AgentID: "agent-1", BuyerRef: "buyer-1"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	for _, input := range []AddItemInput{
		{
			MerchantID:    "merchant-1",
			Domain:        "travel.hotel",
			OfferID:       "offer-jpy",
			OfferSnapshot: json.RawMessage(`{"offer_id":"offer-jpy"}`),
			Amount:        100,
			Currency:      "JPY",
		},
		{
			MerchantID:    "merchant-1",
			Domain:        "travel.hotel",
			OfferID:       "offer-usd",
			OfferSnapshot: json.RawMessage(`{"offer_id":"offer-usd"}`),
			Amount:        100,
			Currency:      "USD",
		},
	} {
		if _, err := service.AddItem(ctx, created.ID, input); err != nil {
			t.Fatalf("AddItem() error = %v", err)
		}
	}

	_, err = service.Checkout(ctx, created.ID)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Checkout() error = %v, want invalid input", err)
	}
}
