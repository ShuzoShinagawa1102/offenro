package retailshoes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type IndexBuilder struct {
	clients merchantClientFactory
}

func NewIndexBuilder(httpClient *http.Client) *IndexBuilder {
	return &IndexBuilder{
		clients: &generatedMerchantClientFactory{httpClient: httpClient},
	}
}

func (b *IndexBuilder) Domain() model.Domain {
	return Domain
}

func (b *IndexBuilder) Build(
	ctx context.Context,
	capability model.MerchantCapability,
) ([]model.DiscoveryIndexEntry, error) {
	client, err := b.clients.NewClient(capability)
	if err != nil {
		return nil, err
	}

	response, err := client.GetRetailShoesCatalogWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("call generated merchant catalog client: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, fmt.Errorf(
			"merchant catalog returned status %d",
			response.StatusCode(),
		)
	}

	countByBrand := make(map[string]int)
	for _, product := range response.JSON200.Products {
		brand := normalizeBrand(product.Brand)
		if brand != "" {
			countByBrand[brand]++
		}
	}

	indexedAt := time.Now().UTC()
	entries := make([]model.DiscoveryIndexEntry, 0, len(countByBrand))
	for brand, count := range countByBrand {
		entries = append(entries, model.DiscoveryIndexEntry{
			Domain:      Domain,
			Dimension:   dimensionBrand,
			Value:       brand,
			MerchantID:  capability.Merchant.ID,
			SupplyCount: count,
			IndexedAt:   indexedAt,
		})
	}

	return entries, nil
}
