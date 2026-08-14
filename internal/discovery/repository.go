package discovery

import (
	"context"
	"sort"
	"sync"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type InMemoryIndexRepository struct {
	mu      sync.RWMutex
	entries []IndexEntry
}

func NewInMemoryIndexRepository() *InMemoryIndexRepository {
	return &InMemoryIndexRepository{}
}

func (r *InMemoryIndexRepository) ReplaceDomain(
	ctx context.Context,
	domain search.Domain,
	entries []IndexEntry,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	next :=
		make(
			[]IndexEntry,
			0,
			len(r.entries)+len(entries),
		)

	for _, entry := range r.entries {
		if entry.Domain == domain {
			continue
		}

		next = append(
			next,
			entry,
		)
	}

	next = append(
		next,
		entries...,
	)

	r.entries = next

	return nil
}

func (r *InMemoryIndexRepository) Find(
	ctx context.Context,
	domain search.Domain,
	dimension string,
	value string,
	limit int,
) ([]IndexEntry, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	result :=
		make(
			[]IndexEntry,
			0,
		)

	for _, entry := range r.entries {

		if entry.Domain != domain {
			continue
		}

		if entry.Dimension != dimension {
			continue
		}

		if entry.Value != value {
			continue
		}

		result = append(
			result,
			entry,
		)
	}

	sort.Slice(
		result,
		func(i, j int) bool {

			if result[i].SupplyCount ==
				result[j].SupplyCount {

				return result[i].MerchantID <
					result[j].MerchantID
			}

			return result[i].SupplyCount >
				result[j].SupplyCount
		},
	)

	if limit > 0 &&
		len(result) > limit {

		result = result[:limit]
	}

	return result, nil
}
