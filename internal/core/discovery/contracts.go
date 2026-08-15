package discovery

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type IndexRepository interface {
	ReplaceDomain(
		ctx context.Context,
		domain model.Domain,
		entries []model.DiscoveryIndexEntry,
	) error

	Find(
		ctx context.Context,
		domain model.Domain,
		dimension string,
		value string,
		limit int,
	) ([]model.DiscoveryIndexEntry, error)
}

type DomainIndexBuilder interface {
	Domain() model.Domain
	Build(
		ctx context.Context,
		merchant model.MerchantCapability,
	) ([]model.DiscoveryIndexEntry, error)
}

type DomainRegistry interface {
	Domains() []model.Domain
	IndexBuilder(domain model.Domain) (DomainIndexBuilder, bool)
}
