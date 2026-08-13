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

// Merchant側ProtocolのRequest。
type travelHotelSearchRequest struct {
	Location string `json:"location"`
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
	Adults   int    `json:"adults"`
	Rooms    int    `json:"rooms"`
	MaxPrice *int64 `json:"max_price,omitempty"`
}

// Merchant側ProtocolのResponse。
type travelHotelSearchResponse struct {
	Offers []travelHotelOffer `json:"offers"`
}

type travelHotelOffer struct {
	OfferID  string `json:"offer_id"`
	Title    string `json:"title"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func (s *TravelHotelSearcher) Search(
	ctx context.Context,
	merchant search.MerchantTarget,
	condition search.SearchCondition,
) ([]search.Offer, error) {

	// travel.hotel Searcherなので、
	// ConditionをTravelHotelConditionとして扱う。
	hotelCondition, ok :=
		condition.(search.TravelHotelCondition)

	if !ok {
		return nil, fmt.Errorf(
			"invalid condition type for domain %s",
			search.DomainTravelHotel,
		)
	}

	requestBody := travelHotelSearchRequest{
		Location: hotelCondition.Location,
		CheckIn: hotelCondition.CheckIn.Format(
			"2006-01-02",
		),
		CheckOut: hotelCondition.CheckOut.Format(
			"2006-01-02",
		),
		Adults:   hotelCondition.Adults,
		Rooms:    hotelCondition.Rooms,
		MaxPrice: hotelCondition.MaxPrice,
	}

	body, err := json.Marshal(
		requestBody,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal merchant search request: %w",
			err,
		)
	}

	url :=
		merchant.BaseURL +
			"/travel/hotel/search"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create merchant request: %w",
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
			"merchant returned unexpected status: %d",
			response.StatusCode,
		)
	}

	var merchantResponse travelHotelSearchResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&merchantResponse); err != nil {

		return nil, fmt.Errorf(
			"decode merchant response: %w",
			err,
		)
	}

	offers := make(
		[]search.Offer,
		0,
		len(merchantResponse.Offers),
	)

	for _, merchantOffer := range merchantResponse.Offers {

		offers = append(
			offers,
			search.Offer{
				// Prototypeでは
				// Merchant ID + Merchant側Offer IDを
				// Protocol側Offer IDとして利用する。
				//
				// 将来的にはOffenro側で
				// 独立したOffer IDを発行してもよい。
				ID: merchant.MerchantID +
					":" +
					merchantOffer.OfferID,

				MerchantID: merchant.MerchantID,
				Domain:     search.DomainTravelHotel,
				Title:      merchantOffer.Title,
				Amount:     merchantOffer.Amount,
				Currency:   merchantOffer.Currency,
			},
		)
	}

	return offers, nil
}
