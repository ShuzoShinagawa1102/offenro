package travelhotel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	domain "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel"
	agentapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/agent"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/model"
)

type Handler struct {
	searcher search.OfferSearcher
}

var _ agentapi.StrictServerInterface = (*Handler)(nil)

func New(searcher search.OfferSearcher) *Handler {
	return &Handler{searcher: searcher}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	handler := agentapi.NewStrictHandler(h, nil)
	agentapi.HandlerFromMux(handler, mux)
}

func (h *Handler) SearchTravelHotels(
	ctx context.Context,
	request agentapi.SearchTravelHotelsRequestObject,
) (agentapi.SearchTravelHotelsResponseObject, error) {
	if request.Body == nil {
		return nil, fmt.Errorf("request body is required")
	}

	offers, err := h.searcher.SearchOffers(
		ctx,
		domain.Domain,
		&domain.Condition{Request: *request.Body},
	)
	if err != nil {
		return nil, err
	}

	apiOffers := make([]modelapi.TravelHotelOffer, 0, len(offers))
	for _, offer := range offers {
		hotelOffer, ok := offer.(*domain.Offer)
		if !ok {
			return nil, fmt.Errorf("unexpected offer type for domain %s", domain.Domain)
		}
		apiOffers = append(apiOffers, hotelOffer.Value)
	}

	response := modelapi.TravelHotelSearchResponse{Offers: apiOffers}
	return agentapi.SearchTravelHotels200JSONResponse(response), nil
}
