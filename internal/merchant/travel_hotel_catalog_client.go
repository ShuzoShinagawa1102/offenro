package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type TravelHotelCatalogItem struct {
	HotelID        string `json:"hotel_id"`
	Name           string `json:"name"`
	PrefectureCode string `json:"prefecture_code"`
	PrefectureName string `json:"prefecture_name"`
	City           string `json:"city"`
}

type travelHotelCatalogResponse struct {
	Hotels []TravelHotelCatalogItem `json:"hotels"`
}

type TravelHotelCatalogClient struct {
	httpClient *http.Client
}

func NewTravelHotelCatalogClient(
	httpClient *http.Client,
) *TravelHotelCatalogClient {

	return &TravelHotelCatalogClient{
		httpClient: httpClient,
	}
}

func (c *TravelHotelCatalogClient) FetchCatalog(
	ctx context.Context,
	target search.MerchantTarget,
) ([]TravelHotelCatalogItem, error) {

	url :=
		target.BaseURL +
			"/travel/hotel/catalog"

	req, err :=
		http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			url,
			nil,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"create catalog request: %w",
			err,
		)
	}

	response, err :=
		c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"call catalog API: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"catalog API returned status %d",
			response.StatusCode,
		)
	}

	var result travelHotelCatalogResponse

	if err :=
		json.NewDecoder(
			response.Body,
		).Decode(&result); err != nil {

		return nil, fmt.Errorf(
			"decode catalog response: %w",
			err,
		)
	}

	return result.Hotels, nil
}
