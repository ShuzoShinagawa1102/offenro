package merchant

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

// Registry resolves merchants that have declared support for a domain.
type Registry interface {
	FindByDomain(ctx context.Context, domain model.Domain) ([]model.MerchantCapability, error)
}

// InMemoryRegistry is the prototype Registry implementation.
type InMemoryRegistry struct {
	capabilities []model.MerchantCapability
}

func NewInMemoryRegistry(capabilities []model.MerchantCapability) *InMemoryRegistry {
	return &InMemoryRegistry{
		capabilities: append([]model.MerchantCapability(nil), capabilities...),
	}
}

func (r *InMemoryRegistry) FindByDomain(
	ctx context.Context,
	domain model.Domain,
) ([]model.MerchantCapability, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	capabilities := make([]model.MerchantCapability, 0)
	for _, capability := range r.capabilities {
		if capability.Domain == domain {
			capabilities = append(capabilities, capability)
		}
	}

	return capabilities, nil
}
