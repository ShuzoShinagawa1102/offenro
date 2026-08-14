package discovery

import (
	"context"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type IndexEntry struct {
	Domain      search.Domain
	Dimension   string
	Value       string
	MerchantID  string
	SupplyCount int
	IndexedAt   time.Time
}

type IndexRepository interface {
	ReplaceDomain(
		ctx context.Context,
		domain search.Domain,
		entries []IndexEntry,
	) error

	Find(
		ctx context.Context,
		domain search.Domain,
		dimension string,
		value string,
		limit int,
	) ([]IndexEntry, error)
}

type DomainIndexBuilder interface {
	Domain() search.Domain

	Build(
		ctx context.Context,
		merchant search.MerchantTarget,
	) ([]IndexEntry, error)
}
