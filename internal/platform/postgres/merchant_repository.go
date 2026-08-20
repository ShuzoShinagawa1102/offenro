package postgres

import (
	"context"
	"fmt"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

var _ merchant.Registry = (*Store)(nil)

func (s *Store) FindByDomain(ctx context.Context, domain model.Domain) ([]model.MerchantCapability, error) {
	rows, err := s.queries.FindActiveMerchantCapabilitiesByDomain(ctx, domain.String())
	if err != nil {
		return nil, fmt.Errorf("find active merchant capabilities: %w", err)
	}

	capabilities := make([]model.MerchantCapability, 0, len(rows))
	for _, row := range rows {
		capabilities = append(capabilities, model.MerchantCapability{
			ID: row.CapabilityID,
			Merchant: model.Merchant{
				ID:      model.MerchantID(row.MerchantID),
				BaseURL: row.ApiBaseUrl,
			},
			Domain: model.Domain(row.DomainID),
		})
	}
	return capabilities, nil
}
