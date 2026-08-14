package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type TravelHotelSearcher struct {
	httpClient *http.Client
}

func NewTravelHotelSearcher(
	httpClient *http.Client,
) *TravelHotelSearcher {

	return &TravelHotelSearcher{
		httpClient: httpClient,
	}
}

type travelHotelSearchRequest struct {
	Destination travelHotelDestination `json:"destination"`
	Stay        travelHotelStay        `json:"stay"`
	Guests      travelHotelGuests      `json:"guests"`
	Filters     travelHotelFilters     `json:"filters,omitempty"`
}

type travelHotelDestination struct {
	PrefectureCode string `json:"prefecture_code"`
}

type travelHotelStay struct {
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
}

type travelHotelGuests struct {
	Adults int `json:"adults"`
	Rooms  int `json:"rooms"`
}

type travelHotelFilters struct {
	MaxPrice *int64 `json:"max_price,omitempty"`
}

type travelHotelSearchResponse struct {
	Offers []travelHotelOffer `json:"offers"`
}

type travelHotelOffer struct {
	OfferID string `json:"offer_id"`

	Hotel travelHotelResponseHotel `json:"hotel"`

	Stay travelHotelStay `json:"stay"`

	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type travelHotelResponseHotel struct {
	HotelID        string `json:"hotel_id"`
	Name           string `json:"name"`
	PrefectureCode string `json:"prefecture_code"`
	PrefectureName string `json:"prefecture_name"`
	City           string `json:"city"`
}

func (s *TravelHotelSearcher) Search(
	ctx context.Context,
	target search.MerchantTarget,
	condition search.SearchCondition,
) ([]search.Offer, error) {

	hotelCondition, ok :=
		condition.(search.TravelHotelCondition)

	if !ok {
		return nil, fmt.Errorf(
			"invalid condition type for domain %s",
			search.DomainTravelHotel,
		)
	}

	requestBody :=
		travelHotelSearchRequest{
			Destination: travelHotelDestination{
				PrefectureCode: hotelCondition.
					Destination.
					PrefectureCode,
			},

			Stay: travelHotelStay{
				CheckIn: hotelCondition.
					Stay.
					CheckIn.
					Format("2006-01-02"),

				CheckOut: hotelCondition.
					Stay.
					CheckOut.
					Format("2006-01-02"),
			},

			Guests: travelHotelGuests{
				Adults: hotelCondition.
					Guests.
					Adults,

				Rooms: hotelCondition.
					Guests.
					Rooms,
			},

			Filters: travelHotelFilters{
				MaxPrice: hotelCondition.
					Filters.
					MaxPrice,
			},
		}

	body, err :=
		json.Marshal(
			requestBody,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal hotel search request: %w",
			err,
		)
	}

	url :=
		target.BaseURL +
			"/travel/hotel/search"

	req, err :=
		http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			url,
			bytes.NewReader(body),
		)
	if err != nil {
		return nil, fmt.Errorf(
			"create merchant search request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	response, err :=
		s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"call merchant search API: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"merchant search returned status %d",
			response.StatusCode,
		)
	}

	var merchantResponse travelHotelSearchResponse

	if err :=
		json.NewDecoder(
			response.Body,
		).Decode(&merchantResponse); err != nil {

		return nil, fmt.Errorf(
			"decode merchant response: %w",
			err,
		)
	}

	offers :=
		make(
			[]search.Offer,
			0,
			len(merchantResponse.Offers),
		)

	for _, merchantOffer := range merchantResponse.Offers {

		offers = append(
			offers,
			search.TravelHotelOffer{
				Base: search.OfferBase{
					ID: target.MerchantID +
						":" +
						merchantOffer.OfferID,

					MerchantID: target.MerchantID,

					Domain: search.DomainTravelHotel,

					Amount: merchantOffer.Amount,

					Currency: merchantOffer.Currency,
				},

				Hotel: search.TravelHotel{
					HotelID: merchantOffer.
						Hotel.
						HotelID,

					Name: merchantOffer.
						Hotel.
						Name,

					PrefectureCode: merchantOffer.
						Hotel.
						PrefectureCode,

					PrefectureName: merchantOffer.
						Hotel.
						PrefectureName,

					City: merchantOffer.
						Hotel.
						City,
				},

				Stay: search.TravelHotelStay{
					CheckIn: hotelCondition.
						Stay.
						CheckIn,

					CheckOut: hotelCondition.
						Stay.
						CheckOut,
				},
			},
		)
	}

	return offers, nil
}
