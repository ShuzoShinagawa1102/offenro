package fulfillment

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

func TestConfirmCreatesMerchantOrderAndConfirmsPurchase(t *testing.T) {
	repository, merchants, registry := fulfillmentFixture(1)
	var submitCalls int
	var receivedKey string
	registry.fulfiller.submit = func(_ context.Context, _ model.MerchantCapability, request MerchantRequest) (MerchantResult, error) {
		submitCalls++
		receivedKey = request.IdempotencyKey
		return MerchantResult{Status: model.MerchantFulfillmentStatusConfirmed, MerchantOrderRef: "order-1"}, nil
	}
	service := NewService(repository, merchants, registry)

	result, err := service.Confirm(context.Background(), ConfirmInput{
		PurchaseID: "purchase-1", Details: map[string]json.RawMessage{"item-1": json.RawMessage(`{"name":"buyer"}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.PurchaseStatusConfirmed || len(result.Fulfillments) != 1 {
		t.Fatalf("result = %#v, want confirmed purchase with one fulfillment", result)
	}
	if submitCalls != 1 || receivedKey == "" || result.Fulfillments[0].IdempotencyKey != receivedKey {
		t.Fatalf("submit calls = %d, key = %q, fulfillment = %#v", submitCalls, receivedKey, result.Fulfillments[0])
	}

	if _, err := service.Confirm(context.Background(), ConfirmInput{PurchaseID: "purchase-1"}); err != nil {
		t.Fatal(err)
	}
	if submitCalls != 1 {
		t.Fatalf("idempotent confirm submitted %d times, want 1", submitCalls)
	}
}

func TestConfirmRecoversUnknownResultWithSameIdempotencyKey(t *testing.T) {
	repository, merchants, registry := fulfillmentFixture(1)
	var submitCalls, resolveCalls int
	var firstKey string
	registry.fulfiller.submit = func(_ context.Context, _ model.MerchantCapability, request MerchantRequest) (MerchantResult, error) {
		submitCalls++
		if firstKey == "" {
			firstKey = request.IdempotencyKey
			return MerchantResult{}, errors.New("response lost")
		}
		if request.IdempotencyKey != firstKey {
			t.Fatalf("retry key = %q, want %q", request.IdempotencyKey, firstKey)
		}
		return MerchantResult{Status: model.MerchantFulfillmentStatusConfirmed, MerchantOrderRef: "order-1"}, nil
	}
	registry.fulfiller.resolve = func(_ context.Context, _ model.MerchantCapability, key string) (MerchantResult, error) {
		resolveCalls++
		if key != firstKey {
			t.Fatalf("resolve key = %q, want %q", key, firstKey)
		}
		return MerchantResult{}, ErrMerchantRecordNotFound
	}
	service := NewService(repository, merchants, registry)

	first, err := service.Confirm(context.Background(), ConfirmInput{
		PurchaseID: "purchase-1", Details: map[string]json.RawMessage{"item-1": json.RawMessage(`{"name":"buyer"}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != model.PurchaseStatusCreated || first.Fulfillments[0].Status != model.MerchantFulfillmentStatusUnknown {
		t.Fatalf("first result = %#v, want CREATED/UNKNOWN", first)
	}

	second, err := service.Confirm(context.Background(), ConfirmInput{PurchaseID: "purchase-1"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Status != model.PurchaseStatusConfirmed || submitCalls != 2 || resolveCalls != 1 {
		t.Fatalf("second status = %s, submit = %d, resolve = %d", second.Status, submitCalls, resolveCalls)
	}
}

func TestConfirmKeepsPurchaseCreatedForPartialResult(t *testing.T) {
	repository, merchants, registry := fulfillmentFixture(2)
	registry.fulfiller.submit = func(_ context.Context, _ model.MerchantCapability, request MerchantRequest) (MerchantResult, error) {
		if request.PurchaseItemID == "item-1" {
			return MerchantResult{Status: model.MerchantFulfillmentStatusConfirmed, MerchantOrderRef: "order-1"}, nil
		}
		return MerchantResult{Status: model.MerchantFulfillmentStatusRejected, FailureCode: "sold_out"}, nil
	}
	service := NewService(repository, merchants, registry)
	result, err := service.Confirm(context.Background(), ConfirmInput{
		PurchaseID: "purchase-1",
		Details: map[string]json.RawMessage{
			"item-1": json.RawMessage(`{"name":"one"}`),
			"item-2": json.RawMessage(`{"name":"two"}`),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.PurchaseStatusCreated {
		t.Fatalf("status = %s, want CREATED for a partial merchant result", result.Status)
	}
}

func TestConfirmRequiresDetailsForEveryNewFulfillment(t *testing.T) {
	repository, merchants, registry := fulfillmentFixture(1)
	service := NewService(repository, merchants, registry)
	_, err := service.Confirm(context.Background(), ConfirmInput{PurchaseID: "purchase-1"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

type fakeRepository struct {
	purchase     model.Purchase
	fulfillments []model.MerchantFulfillment
}

func (r *fakeRepository) FindPurchase(_ context.Context, purchaseID string) (model.Purchase, error) {
	if r.purchase.ID != purchaseID {
		return model.Purchase{}, ErrNotFound
	}
	value := r.purchase
	value.Fulfillments = append([]model.MerchantFulfillment(nil), r.fulfillments...)
	return value, nil
}

func (r *fakeRepository) PrepareFulfillments(_ context.Context, values []model.MerchantFulfillment) error {
	for _, value := range values {
		found := false
		for _, existing := range r.fulfillments {
			found = found || existing.PurchaseItemID == value.PurchaseItemID
		}
		if !found {
			r.fulfillments = append(r.fulfillments, value)
		}
	}
	return nil
}

func (r *fakeRepository) ListFulfillments(_ context.Context, _ string) ([]model.MerchantFulfillment, error) {
	return append([]model.MerchantFulfillment(nil), r.fulfillments...), nil
}

func (r *fakeRepository) UpdateFulfillment(_ context.Context, value model.MerchantFulfillment, expected model.MerchantFulfillmentStatus) error {
	for index := range r.fulfillments {
		if r.fulfillments[index].ID == value.ID && r.fulfillments[index].Status == expected {
			r.fulfillments[index] = value
		}
	}
	return nil
}

func (r *fakeRepository) UpdatePurchaseStatus(_ context.Context, _ string, expected, status model.PurchaseStatus) error {
	if r.purchase.Status == expected {
		r.purchase.Status = status
	}
	return nil
}

type fakeMerchantRegistry struct {
	capabilities map[string]model.MerchantCapability
}

func (r *fakeMerchantRegistry) FindByDomain(_ context.Context, domain model.Domain) ([]model.MerchantCapability, error) {
	var result []model.MerchantCapability
	for _, capability := range r.capabilities {
		if capability.Domain == domain {
			result = append(result, capability)
		}
	}
	return result, nil
}

func (r *fakeMerchantRegistry) FindByID(_ context.Context, capabilityID string) (model.MerchantCapability, error) {
	capability, ok := r.capabilities[capabilityID]
	if !ok {
		return model.MerchantCapability{}, merchant.ErrNotFound
	}
	return capability, nil
}

type fakeFulfiller struct {
	submit  func(context.Context, model.MerchantCapability, MerchantRequest) (MerchantResult, error)
	resolve func(context.Context, model.MerchantCapability, string) (MerchantResult, error)
}

func (f *fakeFulfiller) Domain() model.Domain { return "test.domain" }
func (f *fakeFulfiller) ValidateDetails(details json.RawMessage) error {
	if !json.Valid(details) {
		return errors.New("invalid JSON")
	}
	return nil
}
func (f *fakeFulfiller) Submit(ctx context.Context, capability model.MerchantCapability, request MerchantRequest) (MerchantResult, error) {
	return f.submit(ctx, capability, request)
}
func (f *fakeFulfiller) Resolve(ctx context.Context, capability model.MerchantCapability, key string) (MerchantResult, error) {
	return f.resolve(ctx, capability, key)
}

type fakeFulfillerRegistry struct{ fulfiller *fakeFulfiller }

func (r *fakeFulfillerRegistry) Fulfiller(domain model.Domain) (DomainFulfiller, bool) {
	return r.fulfiller, domain == r.fulfiller.Domain()
}

func fulfillmentFixture(itemCount int) (*fakeRepository, *fakeMerchantRegistry, *fakeFulfillerRegistry) {
	items := make([]model.PurchaseItem, 0, itemCount)
	capabilities := make(map[string]model.MerchantCapability, itemCount)
	for index := 1; index <= itemCount; index++ {
		suffix := strconv.Itoa(index)
		itemID := "item-" + suffix
		capabilityID := "capability-" + suffix
		merchantID := model.MerchantID("merchant-" + suffix)
		items = append(items, model.PurchaseItem{
			ID: itemID, PurchaseID: "purchase-1", CapabilityID: capabilityID,
			MerchantID: merchantID, Domain: "test.domain", MerchantOfferRef: "offer-ref-" + itemID,
		})
		capabilities[capabilityID] = model.MerchantCapability{
			ID: capabilityID, Domain: "test.domain", Merchant: model.Merchant{ID: merchantID},
		}
	}
	repository := &fakeRepository{purchase: model.Purchase{
		ID: "purchase-1", BuyerRef: "buyer-1", Status: model.PurchaseStatusCreated, Items: items,
	}}
	fulfiller := &fakeFulfiller{
		submit: func(context.Context, model.MerchantCapability, MerchantRequest) (MerchantResult, error) {
			return MerchantResult{Status: model.MerchantFulfillmentStatusConfirmed, MerchantOrderRef: "order"}, nil
		},
		resolve: func(context.Context, model.MerchantCapability, string) (MerchantResult, error) {
			return MerchantResult{}, ErrMerchantRecordNotFound
		},
	}
	return repository, &fakeMerchantRegistry{capabilities: capabilities}, &fakeFulfillerRegistry{fulfiller: fulfiller}
}
