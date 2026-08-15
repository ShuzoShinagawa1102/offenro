package search

import (
	"context"
	"errors"
	"testing"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type testCondition struct {
	domain     model.Domain
	validation error
}

func (c testCondition) Domain() model.Domain { return c.domain }
func (c testCondition) Validate() error      { return c.validation }

type testOffer struct {
	domain model.Domain
	id     string
}

func (o testOffer) Domain() model.Domain { return o.domain }

type testDiscovery struct{}

func (testDiscovery) FindMerchants(
	_ context.Context,
	_ SearchCondition,
	candidates []model.MerchantCapability,
	limit int,
) ([]model.MerchantCapability, error) {
	if limit > 0 && len(candidates) > limit {
		return candidates[:limit], nil
	}
	return candidates, nil
}

type testSearcher struct {
	calls  int
	domain model.Domain
}

func (s *testSearcher) Search(
	_ context.Context,
	_ model.MerchantCapability,
	_ SearchCondition,
) ([]Offer, error) {
	s.calls++
	return []Offer{
		testOffer{domain: s.domain, id: "first"},
		testOffer{domain: s.domain, id: "second"},
	}, nil
}

type testDomainRegistry struct {
	domain    model.Domain
	searcher  DomainSearcher
	discovery MerchantDiscovery
}

func (r testDomainRegistry) Searcher(domain model.Domain) (DomainSearcher, bool) {
	return r.searcher, domain == r.domain
}

func (r testDomainRegistry) MerchantDiscovery(domain model.Domain) (MerchantDiscovery, bool) {
	return r.discovery, domain == r.domain
}

func TestServiceSearchOffersUsesRegisteredDomainComponents(t *testing.T) {
	t.Parallel()

	domain := model.Domain("test.product")
	capabilities := []model.MerchantCapability{
		{Merchant: model.Merchant{ID: "merchant-1"}, Domain: domain},
		{Merchant: model.Merchant{ID: "merchant-2"}, Domain: domain},
	}
	domainSearcher := &testSearcher{domain: domain}
	service := NewService(
		merchant.NewInMemoryRegistry(capabilities),
		testDomainRegistry{
			domain:    domain,
			searcher:  domainSearcher,
			discovery: testDiscovery{},
		},
		1,
		1,
	)

	offers, err := service.SearchOffers(
		context.Background(),
		domain,
		testCondition{domain: domain},
	)
	if err != nil {
		t.Fatalf("SearchOffers() error = %v", err)
	}
	if domainSearcher.calls != 1 {
		t.Fatalf("search calls = %d, want 1", domainSearcher.calls)
	}
	if len(offers) != 1 {
		t.Fatalf("offer count = %d, want 1", len(offers))
	}
}

func TestServiceSearchOffersRejectsInvalidCondition(t *testing.T) {
	t.Parallel()

	domain := model.Domain("test.product")
	wantErr := errors.New("invalid test condition")
	service := NewService(
		merchant.NewInMemoryRegistry(nil),
		testDomainRegistry{domain: domain},
		1,
		1,
	)

	_, err := service.SearchOffers(
		context.Background(),
		domain,
		testCondition{domain: domain, validation: wantErr},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("SearchOffers() error = %v, want wrapped %v", err, wantErr)
	}
}
