package search

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

const (
	DefaultMerchantLimit = 10
	DefaultOfferLimit    = 50
)

type Service struct {
	merchants merchant.Registry
	domains   DomainRegistry

	merchantLimit int
	offerLimit    int
}

func NewService(
	merchants merchant.Registry,
	domains DomainRegistry,
	merchantLimit int,
	offerLimit int,
) *Service {
	if merchantLimit <= 0 {
		merchantLimit = DefaultMerchantLimit
	}
	if offerLimit <= 0 {
		offerLimit = DefaultOfferLimit
	}

	return &Service{
		merchants:     merchants,
		domains:       domains,
		merchantLimit: merchantLimit,
		offerLimit:    offerLimit,
	}
}

func (s *Service) SearchOffers(
	ctx context.Context,
	domain model.Domain,
	condition SearchCondition,
) ([]Offer, error) {
	if condition == nil {
		return nil, fmt.Errorf("condition is required")
	}
	if condition.Domain() != domain {
		return nil, fmt.Errorf(
			"domain mismatch: requested=%s condition=%s",
			domain,
			condition.Domain(),
		)
	}
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid search condition: %w", err)
	}

	discovery, ok := s.domains.MerchantDiscovery(domain)
	if !ok {
		return nil, fmt.Errorf("merchant discovery is not registered for domain: %s", domain)
	}
	searcher, ok := s.domains.Searcher(domain)
	if !ok {
		return nil, fmt.Errorf("searcher is not registered for domain: %s", domain)
	}

	candidates, err := s.merchants.FindByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("find merchant capabilities: %w", err)
	}
	if len(candidates) == 0 {
		return []Offer{}, nil
	}

	merchants, err := discovery.FindMerchants(
		ctx,
		condition,
		candidates,
		s.merchantLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("discover merchants: %w", err)
	}

	offers := make([]Offer, 0)
	for _, candidate := range merchants {
		foundOffers, searchErr := searcher.Search(ctx, candidate, condition)
		if searchErr != nil {
			slog.Warn(
				"merchant live search failed",
				"merchant_id", candidate.Merchant.ID,
				"domain", domain,
				"error", searchErr,
			)
			continue
		}

		for _, offer := range foundOffers {
			if offer.Domain() != domain {
				slog.Warn(
					"merchant returned unexpected domain",
					"merchant_id", candidate.Merchant.ID,
					"expected", domain,
					"actual", offer.Domain(),
				)
				continue
			}
			offers = append(offers, offer)
		}
	}

	if len(offers) > s.offerLimit {
		offers = offers[:s.offerLimit]
	}

	return offers, nil
}
