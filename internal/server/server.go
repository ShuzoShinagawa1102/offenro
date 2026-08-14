package server

import (
	"context"
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type Server struct {
	searchService *search.Service
}

var _ api.StrictServerInterface =
	(*Server)(nil)

func New(
	searchService *search.Service,
) *Server {

	return &Server{
		searchService:
			searchService,
	}
}

func (s *Server) GetHealth(
	ctx context.Context,
	request api.GetHealthRequestObject,
) (
	api.GetHealthResponseObject,
	error,
) {

	return api.GetHealth200JSONResponse{
		Status: "ok",
	}, nil
}

func (s *Server) SearchTravelHotels(
	ctx context.Context,
	request api.SearchTravelHotelsRequestObject,
) (
	api.SearchTravelHotelsResponseObject,
	error,
) {

	body := request.Body

	var maxPrice *int64

	if body.Filters != nil {
		maxPrice =
			body.
				Filters.
				MaxPrice
	}

	condition :=
		search.TravelHotelCondition{

			Destination:
				search.TravelHotelDestination{
					PrefectureCode:
						body.
							Destination.
							PrefectureCode,
				},

			Stay:
				search.TravelHotelStay{
					CheckIn:
						body.
							Stay.
							CheckIn.
							Time,

					CheckOut:
						body.
							Stay.
							CheckOut.
							Time,
				},

			Guests:
				search.TravelHotelGuests{
					Adults:
						body.
							Guests.
							Adults,

					Rooms:
						body.
							Guests.
							Rooms,
				},

			Filters:
				search.TravelHotelFilters{
					MaxPrice:
						maxPrice,
				},
		}

	offers, err :=
		s.searchService.SearchOffers(
			ctx,
			search.DomainTravelHotel,
			condition,
		)

	if err != nil {
		return nil, err
	}

	apiOffers :=
		make(
			[]api.TravelHotelOffer,
			0,
			len(offers),
		)

	for _, offer := range offers {

		hotelOffer, ok :=
			offer.(search.TravelHotelOffer)

		if !ok {
			return nil, fmt.Errorf(
				"unexpected offer type for travel.hotel",
			)
		}

		apiOffers = append(
			apiOffers,
			api.TravelHotelOffer{
				OfferId:
					hotelOffer.Base.ID,

				MerchantId:
					hotelOffer.Base.MerchantID,

				Domain:
					string(
						hotelOffer.
							Base.
							Domain,
					),

				Amount:
					hotelOffer.
						Base.
						Amount,

				Currency:
					hotelOffer.
						Base.
						Currency,

				Hotel:
					api.TravelHotelHotel{
						HotelId:
							hotelOffer.
								Hotel.
								HotelID,

						Name:
							hotelOffer.
								Hotel.
								Name,

						PrefectureCode:
							hotelOffer.
								Hotel.
								PrefectureCode,

						PrefectureName:
							hotelOffer.
								Hotel.
								PrefectureName,

						City:
							hotelOffer.
								Hotel.
								City,
					},

				Stay:
					api.TravelHotelStay{
						CheckIn:
							openapi_types.Date{
								Time:
									hotelOffer.
										Stay.
										CheckIn,
							},

						CheckOut:
							openapi_types.Date{
								Time:
									hotelOffer.
										Stay.
										CheckOut,
							},
					},
			},
		)
	}

	return api.SearchTravelHotels200JSONResponse{
		Offers: apiOffers,
	}, nil
}
