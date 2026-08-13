package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

type TravelHotelSearchRequest struct {
	Location string `json:"location"`
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
	Adults   int    `json:"adults"`
	Rooms    int    `json:"rooms"`
	MaxPrice *int64 `json:"max_price,omitempty"`
}

type TravelHotelSearchResponse struct {
	Offers []HotelOffer `json:"offers"`
}

type HotelOffer struct {
	OfferID  string `json:"offer_id"`
	Title    string `json:"title"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /travel/hotel/search",
		searchTravelHotels,
	)

	httpServer := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	slog.Info(
		"merchant mock started",
		"addr",
		":8081",
	)

	if err := httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {

		slog.Error(
			"merchant mock failed",
			"error",
			err,
		)

		os.Exit(1)
	}
}

func searchTravelHotels(
	w http.ResponseWriter,
	r *http.Request,
) {

	var request TravelHotelSearchRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&request); err != nil {

		http.Error(
			w,
			"invalid request",
			http.StatusBadRequest,
		)

		return
	}

	// 現在は固定Offerを返すMock。
	//
	// 実際のMerchantでは、
	// このConditionを利用して
	// Merchant自身のDBや予約システムを検索する。
	offers := []HotelOffer{
		{
			OfferID: "hotel_offer_001",
			Title: request.Location +
				" 温泉ホテル",
			Amount:   48000,
			Currency: "JPY",
		},
		{
			OfferID: "hotel_offer_002",
			Title: request.Location +
				" リゾートホテル",
			Amount:   42000,
			Currency: "JPY",
		},
	}

	// max_priceが指定されていれば
	// Mock側でも簡単に絞り込む。
	if request.MaxPrice != nil {

		filtered :=
			make([]HotelOffer, 0)

		for _, offer := range offers {

			if offer.Amount <= *request.MaxPrice {
				filtered = append(
					filtered,
					offer,
				)
			}
		}

		offers = filtered
	}

	response :=
		TravelHotelSearchResponse{
			Offers: offers,
		}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(
		http.StatusOK,
	)

	if err := json.NewEncoder(
		w,
	).Encode(response); err != nil {

		slog.Error(
			"encode response failed",
			"error",
			err,
		)
	}
}
