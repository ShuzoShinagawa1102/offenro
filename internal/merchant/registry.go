package merchant

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type InMemoryRegistry struct {
	targets []search.MerchantTarget
}

func NewInMemoryRegistry(
	targets []search.MerchantTarget,
) *InMemoryRegistry {
	return &InMemoryRegistry{
		targets: targets,
	}
}

func (r *InMemoryRegistry) FindByDomain(
	ctx context.Context,
	domain search.Domain,
) ([]search.MerchantTarget, error) {

	var result []search.MerchantTarget

	for _, target := range r.targets {
		if target.Domain == domain {
			result = append(
				result,
				target,
			)
		}
	}

	return result, nil
}
