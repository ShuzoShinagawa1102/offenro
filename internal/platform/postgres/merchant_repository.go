package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/jackc/pgx/v5"
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
			ID: row.CapabilityID, APIBaseURL: row.ApiBaseUrl,
			Status: model.CapabilityStatusActive,
			Merchant: model.Merchant{
				ID:     model.MerchantID(row.MerchantID),
				Status: model.MerchantStatusActive,
			},
			Domain: model.Domain(row.DomainID),
		})
	}
	return capabilities, nil
}

func (s *Store) FindByID(ctx context.Context, capabilityID string) (model.MerchantCapability, error) {
	row, err := s.queries.GetActiveMerchantCapability(ctx, capabilityID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.MerchantCapability{}, fmt.Errorf("%w: active merchant capability", merchant.ErrNotFound)
	}
	if err != nil {
		return model.MerchantCapability{}, fmt.Errorf("get active merchant capability: %w", err)
	}
	return model.MerchantCapability{
		ID: row.CapabilityID, Domain: model.Domain(row.DomainID), APIBaseURL: row.ApiBaseUrl,
		Status: model.CapabilityStatus(row.CapabilityStatus), ProtocolVersion: row.ProtocolVersion,
		Merchant: model.Merchant{
			ID: model.MerchantID(row.MerchantID), Name: row.MerchantName,
			Status: model.MerchantStatus(row.MerchantStatus),
		},
	}, nil
}
