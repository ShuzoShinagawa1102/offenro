package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/mockdata"
)

type travelHotelCatalogResponse struct {
	Hotels []travelHotelCatalogItem `json:"hotels"`
}

type travelHotelCatalogItem struct {
	HotelID        string `json:"hotel_id"`
	Name           string `json:"name"`
	PrefectureCode string `json:"prefecture_code"`
	PrefectureName string `json:"prefecture_name"`
	City           string `json:"city"`
}

type travelHotelSearchRequest struct {
	Destination travelHotelDestination `json:"destination"`
	Stay        travelHotelStay        `json:"stay"`
	Guests      travelHotelGuests      `json:"guests"`
	Filters     *travelHotelFilters    `json:"filters,omitempty"`
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

func main() {

	store :=
		mockdata.NewStore()

	mux :=
		http.NewServeMux()

	mux.HandleFunc(
		"GET /merchants/{merchantID}/travel/hotel/catalog",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			handleTravelHotelCatalog(
				store,
				w,
				r,
			)
		},
	)

	mux.HandleFunc(
		"POST /merchants/{merchantID}/travel/hotel/search",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			handleTravelHotelSearch(
				store,
				w,
				r,
			)
		},
	)

	server :=
		&http.Server{
			Addr:    ":8081",
			Handler: mux,
		}

	slog.Info(
		"merchant mock started",
		"addr", ":8081",
		"merchants", 30,
		"merchant_hotel_records", 2400,
	)

	if err :=
		server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {

		slog.Error(
			"merchant mock failed",
			"error", err,
		)

		os.Exit(1)
	}
}

func handleTravelHotelCatalog(
	store *mockdata.Store,
	w http.ResponseWriter,
	r *http.Request,
) {

	merchantID :=
		r.PathValue(
			"merchantID",
		)

	if !store.HasMerchant(
		merchantID,
	) {

		http.Error(
			w,
			"merchant not found",
			http.StatusNotFound,
		)

		return
	}

	hotels :=
		store.Hotels(
			merchantID,
		)

	items :=
		make(
			[]travelHotelCatalogItem,
			0,
			len(hotels),
		)

	for _, hotel := range hotels {

		items = append(
			items,
			travelHotelCatalogItem{
				HotelID: hotel.HotelID,

				Name: hotel.Name,

				PrefectureCode: hotel.PrefectureCode,

				PrefectureName: hotel.PrefectureName,

				City: hotel.City,
			},
		)
	}

	writeJSON(
		w,
		http.StatusOK,
		travelHotelCatalogResponse{
			Hotels: items,
		},
	)
}

func handleTravelHotelSearch(
	store *mockdata.Store,
	w http.ResponseWriter,
	r *http.Request,
) {

	merchantID :=
		r.PathValue(
			"merchantID",
		)

	if !store.HasMerchant(
		merchantID,
	) {

		http.Error(
			w,
			"merchant not found",
			http.StatusNotFound,
		)

		return
	}

	var request travelHotelSearchRequest

	if err :=
		json.NewDecoder(
			r.Body,
		).Decode(&request); err != nil {

		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)

		return
	}

	checkIn, err :=
		time.Parse(
			"2006-01-02",
			request.Stay.CheckIn,
		)
	if err != nil {
		http.Error(
			w,
			"invalid check_in",
			http.StatusBadRequest,
		)

		return
	}

	checkOut, err :=
		time.Parse(
			"2006-01-02",
			request.Stay.CheckOut,
		)
	if err != nil {
		http.Error(
			w,
			"invalid check_out",
			http.StatusBadRequest,
		)

		return
	}

	if !checkOut.After(checkIn) {
		http.Error(
			w,
			"check_out must be after check_in",
			http.StatusBadRequest,
		)

		return
	}

	if request.Guests.Rooms <= 0 {
		http.Error(
			w,
			"rooms must be greater than zero",
			http.StatusBadRequest,
		)

		return
	}

	hotels :=
		store.Hotels(
			merchantID,
		)

	offers :=
		make(
			[]travelHotelOffer,
			0,
		)

	for _, hotel := range hotels {

		if hotel.PrefectureCode !=
			request.
				Destination.
				PrefectureCode {

			continue
		}

		available := true

		var totalAmount int64

		// Checkout日は宿泊日には含めない。
		for date := checkIn; date.Before(checkOut); date = date.AddDate(
			0,
			0,
			1,
		) {

			rooms,
				pricePerRoom :=
				mockdata.AvailabilityFor(
					hotel.HotelID,
					date,
				)

			if rooms <
				request.Guests.Rooms {

				available = false

				break
			}

			totalAmount +=
				pricePerRoom *
					int64(
						request.
							Guests.
							Rooms,
					)
		}

		if !available {
			continue
		}

		if request.Filters != nil &&
			request.Filters.MaxPrice != nil &&
			totalAmount >
				*request.Filters.MaxPrice {

			continue
		}

		offerID :=
			fmt.Sprintf(
				"%s-%s-%s-r%d",
				hotel.HotelID,
				checkIn.Format(
					"20060102",
				),
				checkOut.Format(
					"20060102",
				),
				request.Guests.Rooms,
			)

		offers = append(
			offers,
			travelHotelOffer{
				OfferID: offerID,

				Hotel: travelHotelResponseHotel{
					HotelID: hotel.HotelID,

					Name: hotel.Name,

					PrefectureCode: hotel.PrefectureCode,

					PrefectureName: hotel.PrefectureName,

					City: hotel.City,
				},

				Stay: request.Stay,

				Amount: totalAmount,

				Currency: "JPY",
			},
		)
	}

	// Mockでは価格の安い順。
	sort.Slice(
		offers,
		func(
			i int,
			j int,
		) bool {

			return offers[i].Amount <
				offers[j].Amount
		},
	)

	writeJSON(
		w,
		http.StatusOK,
		travelHotelSearchResponse{
			Offers: offers,
		},
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(
		status,
	)

	if err :=
		json.NewEncoder(
			w,
		).Encode(value); err != nil {

		slog.Error(
			"encode response failed",
			"error", err,
		)
	}
}
