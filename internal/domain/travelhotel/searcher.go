package travelhotel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
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
	if capability.APIBaseURL == "" {
		return nil, fmt.Errorf("merchant %s has an empty base URL", capability.Merchant.ID)
	}

	client, err := merchantapi.NewClientWithResponses(
		capability.APIBaseURL,
		merchantapi.WithHTTPClient(f.httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("create generated merchant client: %w", err)
	}

	return client, nil
}

type Searcher struct {
	clients merchantClientFactory
	issuer  offer.Issuer
}

func NewSearcher(httpClient *http.Client, issuer offer.Issuer) *Searcher {
	return &Searcher{
		clients: &generatedMerchantClientFactory{httpClient: httpClient},
		issuer:  issuer,
	}
}

func (s *Searcher) Domain() model.Domain { return Domain }

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
		if s.issuer == nil {
			return nil, fmt.Errorf("offer issuer is unavailable")
		}
		expiresAt := time.Time{}
		if merchantOffer.OfferExpiresAt != nil {
			expiresAt = merchantOffer.OfferExpiresAt.UTC()
		}
		issued, err := s.issuer.Issue(offer.Reference{
			CapabilityID: capability.ID, MerchantID: capability.Merchant.ID,
			Domain: Domain, MerchantOfferRef: merchantOffer.MerchantOfferRef,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return nil, fmt.Errorf("issue offer id: %w", err)
		}
		offers = append(offers, &Offer{Value: modelapi.TravelHotelOffer{
			Amount:     merchantOffer.Amount,
			Currency:   merchantOffer.Currency,
			Domain:     string(Domain),
			Hotel:      merchantOffer.Hotel,
			MerchantId: string(capability.Merchant.ID),
			OfferId:    issued.OfferID, OfferExpiresAt: issued.ExpiresAt,
			Stay: merchantOffer.Stay,
		}})
	}

	return offers, nil
}

func (s *Searcher) Revalidate(
	ctx context.Context,
	capability model.MerchantCapability,
	offerID string,
	merchantOfferRef string,
) (offer.Verified, error) {
	if strings.TrimSpace(merchantOfferRef) == "" {
		return offer.Verified{}, offer.ErrInvalidReference
	}
	client, err := s.clients.NewClient(capability)
	if err != nil {
		return offer.Verified{}, err
	}
	response, err := client.RevalidateTravelHotelOfferWithResponse(ctx, modelapi.TravelHotelRevalidateRequest{
		MerchantOfferRef: merchantOfferRef,
	})
	if err != nil {
		return offer.Verified{}, fmt.Errorf("call generated merchant revalidate client: %w", err)
	}
	if response.StatusCode() == http.StatusNotFound {
		return offer.Verified{}, offer.ErrUnavailable
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return offer.Verified{}, fmt.Errorf("merchant revalidate returned status %d", response.StatusCode())
	}
	merchantOffer := response.JSON200.Offer
	if merchantOffer.MerchantOfferRef != merchantOfferRef {
		return offer.Verified{}, fmt.Errorf("%w: merchant returned a different offer reference", offer.ErrUnavailable)
	}
	if merchantOffer.OfferExpiresAt != nil && !merchantOffer.OfferExpiresAt.After(time.Now()) {
		return offer.Verified{}, offer.ErrUnavailable
	}
	agentOffer := modelapi.TravelHotelOffer{
		Amount: merchantOffer.Amount, Currency: merchantOffer.Currency, Domain: Domain.String(),
		Hotel: merchantOffer.Hotel, MerchantId: string(capability.Merchant.ID), OfferId: offerID,
		Stay: merchantOffer.Stay,
	}
	if merchantOffer.OfferExpiresAt != nil {
		agentOffer.OfferExpiresAt = merchantOffer.OfferExpiresAt.UTC()
	}
	snapshot, err := json.Marshal(agentOffer)
	if err != nil {
		return offer.Verified{}, fmt.Errorf("encode revalidated offer: %w", err)
	}
	return offer.Verified{
		Snapshot: snapshot, Amount: merchantOffer.Amount, Currency: merchantOffer.Currency,
		ExpiresAt: merchantOffer.OfferExpiresAt,
	}, nil
}
