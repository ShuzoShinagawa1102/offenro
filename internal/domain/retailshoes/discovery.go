package retailshoes

import (
	"context"
	"fmt"
	"strings"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
)

type MerchantDiscovery struct {
	repository discovery.IndexRepository
}

func NewMerchantDiscovery(repository discovery.IndexRepository) *MerchantDiscovery {
	return &MerchantDiscovery{repository: repository}
}

func (d *MerchantDiscovery) FindMerchants(
	ctx context.Context,
	condition search.SearchCondition,
	candidates []model.MerchantCapability,
	limit int,
) ([]model.MerchantCapability, error) {
	shoesCondition, err := asCondition(condition)
	if err != nil {
		return nil, err
	}

	entries, err := d.repository.Find(
		ctx,
		Domain,
		dimensionBrand,
		normalizeBrand(shoesCondition.Request.Criteria.Brand),
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("find retail.shoes discovery entries: %w", err)
	}

	candidateByMerchant := make(
		map[model.MerchantID]model.MerchantCapability,
		len(candidates),
	)
	for _, candidate := range candidates {
		candidateByMerchant[candidate.Merchant.ID] = candidate
	}

	result := make([]model.MerchantCapability, 0)
	for _, entry := range entries {
		candidate, exists := candidateByMerchant[entry.MerchantID]
		if !exists {
			continue
		}
		result = append(result, candidate)
		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result, nil
}

func normalizeBrand(brand string) string {
	return strings.ToLower(strings.TrimSpace(brand))
}
