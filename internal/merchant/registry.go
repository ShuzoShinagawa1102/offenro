package merchant

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type InMemoryCapabilityRegistry struct {
	targets []search.MerchantTarget
}

func NewInMemoryCapabilityRegistry(
	targets []search.MerchantTarget,
) *InMemoryCapabilityRegistry {

	return &InMemoryCapabilityRegistry{
		targets: targets,
	}
}

func (r *InMemoryCapabilityRegistry) FindByDomain(
	ctx context.Context,
	domain search.Domain,
) ([]search.MerchantTarget, error) {

	result :=
		make(
			[]search.MerchantTarget,
			0,
		)

	for _, target := range r.targets {
		if target.Domain != domain {
			continue
		}

		result = append(
			result,
			target,
		)
	}

	return result, nil
}
