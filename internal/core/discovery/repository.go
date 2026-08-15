package discovery

import (
	"context"
	"sort"
	"sync"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type InMemoryIndexRepository struct {
	mu      sync.RWMutex
	entries []model.DiscoveryIndexEntry
}

func NewInMemoryIndexRepository() *InMemoryIndexRepository {
	return &InMemoryIndexRepository{}
}

func (r *InMemoryIndexRepository) ReplaceDomain(
	ctx context.Context,
	domain model.Domain,
	entries []model.DiscoveryIndexEntry,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	next := make([]model.DiscoveryIndexEntry, 0, len(r.entries)+len(entries))
	for _, entry := range r.entries {
		if entry.Domain != domain {
			next = append(next, entry)
		}
	}
	next = append(next, entries...)
	r.entries = next

	return nil
}

func (r *InMemoryIndexRepository) Find(
	ctx context.Context,
	domain model.Domain,
	dimension string,
	value string,
	limit int,
) ([]model.DiscoveryIndexEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]model.DiscoveryIndexEntry, 0)
	for _, entry := range r.entries {
		if entry.Domain == domain && entry.Dimension == dimension && entry.Value == value {
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].SupplyCount == entries[j].SupplyCount {
			return entries[i].MerchantID < entries[j].MerchantID
		}
		return entries[i].SupplyCount > entries[j].SupplyCount
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}
