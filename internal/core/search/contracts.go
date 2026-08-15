package search

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

// SearchCondition is implemented by a domain-specific condition adapter.
type SearchCondition interface {
	Domain() model.Domain
	Validate() error
}

// Offer is implemented by a domain-specific offer adapter.
type Offer interface {
	Domain() model.Domain
}

// DomainSearcher performs a live search against a merchant API.
type DomainSearcher interface {
	Search(
		ctx context.Context,
		merchant model.MerchantCapability,
		condition SearchCondition,
	) ([]Offer, error)
}

// MerchantDiscovery selects the merchants most likely to satisfy a condition.
type MerchantDiscovery interface {
	FindMerchants(
		ctx context.Context,
		condition SearchCondition,
		candidates []model.MerchantCapability,
		limit int,
	) ([]model.MerchantCapability, error)
}

// DomainRegistry resolves search components registered by Domain Extensions.
type DomainRegistry interface {
	Searcher(domain model.Domain) (DomainSearcher, bool)
	MerchantDiscovery(domain model.Domain) (MerchantDiscovery, bool)
}

// OfferSearcher is the domain-independent SearchOffers entry point.
type OfferSearcher interface {
	SearchOffers(
		ctx context.Context,
		domain model.Domain,
		condition SearchCondition,
	) ([]Offer, error)
}
