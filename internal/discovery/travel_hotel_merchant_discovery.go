package discovery

import (
	"context"
	"fmt"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type TravelHotelMerchantDiscovery struct {
	repository IndexRepository
}

func NewTravelHotelMerchantDiscovery(
	repository IndexRepository,
) *TravelHotelMerchantDiscovery {

	return &TravelHotelMerchantDiscovery{
		repository: repository,
	}
}

func (d *TravelHotelMerchantDiscovery) FindMerchants(
	ctx context.Context,
	condition search.SearchCondition,
	candidates []search.MerchantTarget,
	limit int,
) ([]search.MerchantTarget, error) {

	hotelCondition, ok :=
		condition.(search.TravelHotelCondition)

	if !ok {
		return nil, fmt.Errorf(
			"invalid condition type for travel.hotel discovery",
		)
	}

	entries, err :=
		d.repository.Find(
			ctx,
			search.DomainTravelHotel,
			TravelHotelDimensionPrefectureCode,
			hotelCondition.
				Destination.
				PrefectureCode,

			0,
		)
	if err != nil {
		return nil, err
	}

	candidateMap :=
		make(
			map[string]search.MerchantTarget,
			len(candidates),
		)

	for _, candidate := range candidates {
		candidateMap[candidate.MerchantID] = candidate
	}

	result :=
		make(
			[]search.MerchantTarget,
			0,
		)

	for _, entry := range entries {

		target, exists :=
			candidateMap[entry.MerchantID]

		if !exists {
			continue
		}

		result = append(
			result,
			target,
		)

		if limit > 0 &&
			len(result) >= limit {
			break
		}
	}

	return result, nil
}
