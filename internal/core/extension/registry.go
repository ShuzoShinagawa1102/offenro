package extension

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
)

// Extension connects all domain-specific adapters to Protocol Core.
type Extension interface {
	Domain() model.Domain
	Searcher() search.DomainSearcher
	MerchantDiscovery() search.MerchantDiscovery
	IndexBuilder() discovery.DomainIndexBuilder
}

type Registry struct {
	mu         sync.RWMutex
	extensions map[model.Domain]Extension
}

func NewRegistry() *Registry {
	return &Registry{extensions: make(map[model.Domain]Extension)}
}

func (r *Registry) Register(extension Extension) error {
	if extension == nil {
		return fmt.Errorf("domain extension is required")
	}

	domain := extension.Domain()
	if domain == "" {
		return fmt.Errorf("domain extension has an empty domain")
	}
	if extension.Searcher() == nil {
		return fmt.Errorf("domain extension %s has no searcher", domain)
	}
	if extension.MerchantDiscovery() == nil {
		return fmt.Errorf("domain extension %s has no merchant discovery", domain)
	}
	if extension.IndexBuilder() == nil {
		return fmt.Errorf("domain extension %s has no index builder", domain)
	}
	if extension.IndexBuilder().Domain() != domain {
		return fmt.Errorf("domain extension %s has a mismatched index builder", domain)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.extensions[domain]; exists {
		return fmt.Errorf("domain extension is already registered: %s", domain)
	}
	r.extensions[domain] = extension

	return nil
}

func (r *Registry) Domains() []model.Domain {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domains := make([]model.Domain, 0, len(r.extensions))
	for domain := range r.extensions {
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool {
		return domains[i] < domains[j]
	})

	return domains
}

func (r *Registry) Searcher(domain model.Domain) (search.DomainSearcher, bool) {
	extension, ok := r.find(domain)
	if !ok {
		return nil, false
	}
	return extension.Searcher(), true
}

func (r *Registry) MerchantDiscovery(domain model.Domain) (search.MerchantDiscovery, bool) {
	extension, ok := r.find(domain)
	if !ok {
		return nil, false
	}
	return extension.MerchantDiscovery(), true
}

func (r *Registry) IndexBuilder(domain model.Domain) (discovery.DomainIndexBuilder, bool) {
	extension, ok := r.find(domain)
	if !ok {
		return nil, false
	}
	return extension.IndexBuilder(), true
}

func (r *Registry) find(domain model.Domain) (Extension, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	extension, ok := r.extensions[domain]
	return extension, ok
}
