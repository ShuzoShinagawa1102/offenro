package testutil

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type MerchantRegistry struct {
	capabilities []model.MerchantCapability
}

var _ merchant.Registry = (*MerchantRegistry)(nil)

func NewMerchantRegistry(capabilities []model.MerchantCapability) *MerchantRegistry {
	return &MerchantRegistry{
		capabilities: append([]model.MerchantCapability(nil), capabilities...),
	}
}

func (r *MerchantRegistry) FindByDomain(
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

func (r *MerchantRegistry) FindByID(
	ctx context.Context,
	capabilityID string,
) (model.MerchantCapability, error) {
	if err := ctx.Err(); err != nil {
		return model.MerchantCapability{}, err
	}
	for _, capability := range r.capabilities {
		if capability.ID == capabilityID {
			return capability, nil
		}
	}
	return model.MerchantCapability{}, merchant.ErrNotFound
}
