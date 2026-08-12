package server

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
)

type Server struct{}

var _ api.StrictServerInterface = (*Server)(nil)

func New() *Server {
	return &Server{}
}

func (s *Server) GetHealth(
	ctx context.Context,
	request api.GetHealthRequestObject,
) (api.GetHealthResponseObject, error) {

	return api.GetHealth200JSONResponse{
		Status: "ok",
	}, nil
}

func (s *Server) SearchOffers(
	ctx context.Context,
	request api.SearchOffersRequestObject,
) (api.SearchOffersResponseObject, error) {

	body := request.Body

	offers := []api.Offer{
		{
			OfferId:    "off_001",
			MerchantId: "merchant_001",
			Domain:     body.Domain,
			Title:      "Sample Offer",
			Amount:     1000,
			Currency:   "JPY",
		},
	}

	return api.SearchOffers200JSONResponse{
		Offers: offers,
	}, nil
}
