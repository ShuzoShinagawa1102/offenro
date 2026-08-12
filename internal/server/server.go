package server

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
	"github.com/ShuzoShinagawa1102/offenro/internal/merchant"
)

type Server struct {
	merchantClient *merchant.Client
}

var _ api.StrictServerInterface = (*Server)(nil)

func New(merchantClient *merchant.Client) *Server {
	return &Server{
		merchantClient: merchantClient,
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

func (s *Server) SearchOffers(
	ctx context.Context,
	request api.SearchOffersRequestObject,
) (api.SearchOffersResponseObject, error) {

	// Merchantへ問い合わせる
	result, err := s.merchantClient.SearchShoes(
		ctx,
		merchant.SearchRequest{
			Query: request.Body.Query,
		},
	)
	if err != nil {
		return nil, err
	}

	// Merchant問い合わせ結果をプロトコル用Offerに変換する
	offers := make([]api.Offer, 0, len(result.Offers))
	for _, offer := range result.Offers {
		offers = append(offers, api.Offer{
			OfferId:    offer.OfferID,
			MerchantId: "merchant_001",
			Domain:     request.Body.Domain,
			Title:      offer.Title,
			Amount:     offer.Amount,
			Currency:   offer.Currency,
		})
	}

	return api.SearchOffers200JSONResponse{
		Offers: offers,
	}, nil
}
