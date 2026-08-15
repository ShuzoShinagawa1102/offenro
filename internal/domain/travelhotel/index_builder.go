package travelhotel

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

	response, err := client.GetTravelHotelCatalogWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("call generated merchant catalog client: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, fmt.Errorf(
			"merchant catalog returned status %d",
			response.StatusCode(),
		)
	}

	countByPrefecture := make(map[string]int)
	for _, hotel := range response.JSON200.Hotels {
		if hotel.PrefectureCode != "" {
			countByPrefecture[hotel.PrefectureCode]++
		}
	}

	indexedAt := time.Now().UTC()
	entries := make([]model.DiscoveryIndexEntry, 0, len(countByPrefecture))
	for prefectureCode, count := range countByPrefecture {
		entries = append(entries, model.DiscoveryIndexEntry{
			Domain:      Domain,
			Dimension:   dimensionPrefectureCode,
			Value:       prefectureCode,
			MerchantID:  capability.Merchant.ID,
			SupplyCount: count,
			IndexedAt:   indexedAt,
		})
	}

	return entries, nil
}
