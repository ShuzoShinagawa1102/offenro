package cart

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
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
	r.cart.Items = append(r.cart.Items, item)
	r.cart.UpdatedAt = updatedAt
	return item, nil
}

type testOfferResolver struct {
	references map[string]offer.Reference
}

func (r *testOfferResolver) Resolve(offerID string) (offer.Reference, error) {
	value, ok := r.references[offerID]
	if !ok {
		return offer.Reference{}, offer.ErrInvalidReference
	}
	return value, nil
}

type testMerchantRegistry struct {
	capability model.MerchantCapability
}

func (r *testMerchantRegistry) FindByDomain(context.Context, model.Domain) ([]model.MerchantCapability, error) {
	return []model.MerchantCapability{r.capability}, nil
}

func (r *testMerchantRegistry) FindByID(context.Context, string) (model.MerchantCapability, error) {
	return r.capability, nil
}

type testVerifier struct {
	values map[string]offer.Verified
}

func (v *testVerifier) Domain() model.Domain { return model.Domain("travel.hotel") }

func (v *testVerifier) Revalidate(_ context.Context, _ model.MerchantCapability, _ string, merchantOfferRef string) (offer.Verified, error) {
	value, ok := v.values[merchantOfferRef]
	if !ok {
		return offer.Verified{}, offer.ErrUnavailable
	}
	return value, nil
}

type testVerifierRegistry struct{ verifier offer.Verifier }

func (r *testVerifierRegistry) OfferVerifier(domain model.Domain) (offer.Verifier, bool) {
	return r.verifier, domain == r.verifier.Domain()
}

func newTestService(repository *testRepository, offerIDs map[string]string, verified map[string]offer.Verified) *Service {
	expiresAt := time.Now().Add(time.Hour).UTC()
	references := make(map[string]offer.Reference, len(offerIDs))
	for offerID, merchantOfferRef := range offerIDs {
		references[offerID] = offer.Reference{
			CapabilityID: "capability-1", MerchantID: model.MerchantID("merchant-1"),
			Domain: model.Domain("travel.hotel"), MerchantOfferRef: merchantOfferRef,
			ExpiresAt: expiresAt,
		}
	}
	capability := model.MerchantCapability{
		ID: "capability-1", Domain: model.Domain("travel.hotel"), Status: model.CapabilityStatusActive,
		Merchant: model.Merchant{ID: model.MerchantID("merchant-1"), Status: model.MerchantStatusActive},
	}
	verifier := &testVerifier{values: verified}
	return NewService(
		repository, repository, &testOfferResolver{references: references},
		&testMerchantRegistry{capability: capability}, &testVerifierRegistry{verifier: verifier},
	)
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
	service := newTestService(repository, map[string]string{"offer-1": "merchant-offer-1"}, map[string]offer.Verified{
		"merchant-offer-1": {Snapshot: json.RawMessage(`{"offer_id":"offer-1","revalidated":true}`), Amount: 12000, Currency: "JPY"},
	})
	ctx := context.Background()

	created, err := service.Create(ctx, CreateInput{
		AgentID:  "agent-1",
		BuyerRef: "buyer-1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	item, err := service.AddItem(ctx, created.ID, AddItemInput{
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
	service := newTestService(repository, map[string]string{
		"offer-jpy": "merchant-offer-jpy", "offer-usd": "merchant-offer-usd",
	}, map[string]offer.Verified{
		"merchant-offer-jpy": {Snapshot: json.RawMessage(`{"offer_id":"offer-jpy"}`), Amount: 100, Currency: "JPY"},
		"merchant-offer-usd": {Snapshot: json.RawMessage(`{"offer_id":"offer-usd"}`), Amount: 100, Currency: "USD"},
	})
	ctx := context.Background()
	created, err := service.Create(ctx, CreateInput{AgentID: "agent-1", BuyerRef: "buyer-1"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	for _, input := range []AddItemInput{
		{
			OfferID:       "offer-jpy",
			OfferSnapshot: json.RawMessage(`{"offer_id":"offer-jpy"}`),
			Amount:        100,
			Currency:      "JPY",
		},
		{
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

func TestServiceCheckoutRejectsChangedMerchantPrice(t *testing.T) {
	t.Parallel()

	repository := &testRepository{}
	service := newTestService(repository, map[string]string{"offer-1": "merchant-offer-1"}, map[string]offer.Verified{
		"merchant-offer-1": {Snapshot: json.RawMessage(`{"offer_id":"offer-1"}`), Amount: 13000, Currency: "JPY"},
	})
	ctx := context.Background()
	created, err := service.Create(ctx, CreateInput{AgentID: "agent-1", BuyerRef: "buyer-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddItem(ctx, created.ID, AddItemInput{
		OfferID: "offer-1", OfferSnapshot: json.RawMessage(`{"offer_id":"offer-1"}`),
		Amount: 12000, Currency: "JPY",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Checkout(ctx, created.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("Checkout() error = %v, want conflict", err)
	}
	if repository.cart.Status != model.CartStatusActive {
		t.Fatalf("cart status = %s, want ACTIVE", repository.cart.Status)
	}
}
