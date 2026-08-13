package server

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type Server struct {
	searchService *search.Service
}

var _ api.StrictServerInterface = (*Server)(nil)

func New(
	searchService *search.Service,
) *Server {
	return &Server{
		searchService: searchService,
	}
}

func (s *Server) GetHealth(
	ctx context.Context,
	request api.GetHealthRequestObject,
) (api.GetHealthResponseObject, error) {

	return api.GetHealth200JSONResponse{
		Status: "ok",
	}, nil
}

func (s *Server) SearchTravelHotels(
	ctx context.Context,
	request api.SearchTravelHotelsRequestObject,
) (api.SearchTravelHotelsResponseObject, error) {

	body := request.Body

	// Web APIのRequestを
	// Protocol CoreのTravelHotelConditionへ変換する。
	condition := search.TravelHotelCondition{
		Location: body.Location,

		// OpenAPI format: date は
		// openapi_types.Dateとして生成され、
		// 内部にtime.Timeを保持している。
		CheckIn: body.CheckIn.Time,

		CheckOut: body.CheckOut.Time,

		Adults: body.Adults,
		Rooms:  body.Rooms,

		MaxPrice: body.MaxPrice,
	}

	// Protocol Coreを呼び出す。
	offers, err := s.searchService.SearchOffers(
		ctx,
		search.DomainTravelHotel,
		condition,
	)
	if err != nil {
		return nil, err
	}

	// Protocol内部Offerから
	// 外部Web API用Offerへ変換する。
	apiOffers := make(
		[]api.Offer,
		0,
		len(offers),
	)

	for _, offer := range offers {
		apiOffers = append(
			apiOffers,
			api.Offer{
				OfferId:    offer.ID,
				MerchantId: offer.MerchantID,
				Domain:     string(offer.Domain),
				Title:      offer.Title,
				Amount:     offer.Amount,
				Currency:   offer.Currency,
			},
		)
	}

	response := api.SearchOffersResponse{
		Offers: apiOffers,
	}

	return api.SearchTravelHotels200JSONResponse(
		response,
	), nil
}
