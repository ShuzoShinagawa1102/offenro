package discovery

import (
	"context"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

const (
	TravelHotelDimensionPrefectureCode =
		"prefecture_code"
)

type TravelHotelIndexBuilder struct {
	catalogClient *merchant.TravelHotelCatalogClient
}

func NewTravelHotelIndexBuilder(
	catalogClient *merchant.TravelHotelCatalogClient,
) *TravelHotelIndexBuilder {

	return &TravelHotelIndexBuilder{
		catalogClient: catalogClient,
	}
}

func (b *TravelHotelIndexBuilder) Domain() search.Domain {
	return search.DomainTravelHotel
}

func (b *TravelHotelIndexBuilder) Build(
	ctx context.Context,
	target search.MerchantTarget,
) ([]IndexEntry, error) {

	hotels, err :=
		b.catalogClient.FetchCatalog(
			ctx,
			target,
		)
	if err != nil {
		return nil, err
	}

	countByPrefecture :=
		make(
			map[string]int,
		)

	for _, hotel := range hotels {

		if hotel.PrefectureCode == "" {
			continue
		}

		countByPrefecture[hotel.PrefectureCode]++
	}

	indexedAt := time.Now().UTC()

	entries :=
		make(
			[]IndexEntry,
			0,
			len(countByPrefecture),
		)

	for prefectureCode, count :=
		range countByPrefecture {

		entries = append(
			entries,
			IndexEntry{
				Domain:
					search.DomainTravelHotel,

				Dimension:
					TravelHotelDimensionPrefectureCode,

				Value:
					prefectureCode,

				MerchantID:
					target.MerchantID,

				SupplyCount:
					count,

				IndexedAt:
					indexedAt,
			},
		)
	}

	return entries, nil
}
