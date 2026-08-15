package travelhotel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	merchantapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/merchant"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/model"
)

type merchantClientFactory interface {
	NewClient(capability model.MerchantCapability) (merchantapi.ClientWithResponsesInterface, error)
}

type generatedMerchantClientFactory struct {
	httpClient *http.Client
}

func (f *generatedMerchantClientFactory) NewClient(
	capability model.MerchantCapability,
) (merchantapi.ClientWithResponsesInterface, error) {
	if capability.Merchant.BaseURL == "" {
		return nil, fmt.Errorf("merchant %s has an empty base URL", capability.Merchant.ID)
	}

	client, err := merchantapi.NewClientWithResponses(
		capability.Merchant.BaseURL,
		merchantapi.WithHTTPClient(f.httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("create generated merchant client: %w", err)
	}

	return client, nil
}

type Searcher struct {
	clients merchantClientFactory
}

func NewSearcher(httpClient *http.Client) *Searcher {
	return &Searcher{
		clients: &generatedMerchantClientFactory{httpClient: httpClient},
	}
}

func (s *Searcher) Search(
	ctx context.Context,
	capability model.MerchantCapability,
	condition search.SearchCondition,
) ([]search.Offer, error) {
	hotelCondition, err := asCondition(condition)
	if err != nil {
		return nil, err
	}

	client, err := s.clients.NewClient(capability)
	if err != nil {
		return nil, err
	}

	response, err := client.SearchTravelHotelOffersWithResponse(
		ctx,
		hotelCondition.Request,
	)
	if err != nil {
		return nil, fmt.Errorf("call generated merchant search client: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, fmt.Errorf(
			"merchant search returned status %d",
			response.StatusCode(),
		)
	}

	offers := make([]search.Offer, 0, len(response.JSON200.Offers))
	for _, merchantOffer := range response.JSON200.Offers {
		offers = append(offers, &Offer{Value: modelapi.TravelHotelOffer{
			Amount:     merchantOffer.Amount,
			Currency:   merchantOffer.Currency,
			Domain:     string(Domain),
			Hotel:      merchantOffer.Hotel,
			MerchantId: string(capability.Merchant.ID),
			OfferId: fmt.Sprintf(
				"%s:%s",
				capability.Merchant.ID,
				merchantOffer.OfferId,
			),
			Stay: merchantOffer.Stay,
		}})
	}

	return offers, nil
}
