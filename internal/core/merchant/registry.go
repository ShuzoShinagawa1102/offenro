package merchant

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

// Registry resolves active merchants that have declared support for a domain.
type Registry interface {
	FindByDomain(ctx context.Context, domain model.Domain) ([]model.MerchantCapability, error)
}
