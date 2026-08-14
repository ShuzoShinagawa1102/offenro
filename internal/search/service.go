package search

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	DefaultMerchantLimit = 10
	DefaultOfferLimit    = 50
)

type Service struct {
	registry    MerchantRegistry
	discoveries map[Domain]MerchantDiscovery
	searchers   map[Domain]DomainSearcher

	merchantLimit int
	offerLimit    int
}

func NewService(
	registry MerchantRegistry,
	discoveries map[Domain]MerchantDiscovery,
	searchers map[Domain]DomainSearcher,
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
		registry:      registry,
		discoveries:   discoveries,
		searchers:     searchers,
		merchantLimit: merchantLimit,
		offerLimit:    offerLimit,
	}
}

func (s *Service) SearchOffers(
	ctx context.Context,
	domain Domain,
	condition SearchCondition,
) ([]Offer, error) {

	if condition == nil {
		return nil, fmt.Errorf(
			"condition is required",
		)
	}

	if condition.Domain() != domain {
		return nil, fmt.Errorf(
			"domain mismatch: requested=%s condition=%s",
			domain,
			condition.Domain(),
		)
	}

	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf(
			"invalid search condition: %w",
			err,
		)
	}

	// まずDomainに対応しているMerchantを取得する。
	candidates, err :=
		s.registry.FindByDomain(
			ctx,
			domain,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"find merchant capabilities: %w",
			err,
		)
	}

	if len(candidates) == 0 {
		return []Offer{}, nil
	}

	// Domain固有のMerchant Discoveryを取得。
	discovery, ok :=
		s.discoveries[domain]
	if !ok {
		return nil, fmt.Errorf(
			"merchant discovery is not registered for domain: %s",
			domain,
		)
	}

	// Conditionを利用して、
	// 問い合わせる価値が高いMerchantだけを選ぶ。
	merchants, err :=
		discovery.FindMerchants(
			ctx,
			condition,
			candidates,
			s.merchantLimit,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"discover merchants: %w",
			err,
		)
	}

	searcher, ok :=
		s.searchers[domain]
	if !ok {
		return nil, fmt.Errorf(
			"searcher is not registered for domain: %s",
			domain,
		)
	}

	var offers []Offer

	// TODO:
	// 現在は逐次HTTPリクエスト。
	// 将来的にはMerchant数が増えた場合、
	// 並列実行 + concurrency limitを導入する。
	for _, merchant := range merchants {

		foundOffers, err :=
			searcher.Search(
				ctx,
				merchant,
				condition,
			)

		if err != nil {
			slog.Warn(
				"merchant live search failed",
				"merchant_id", merchant.MerchantID,
				"domain", domain,
				"error", err,
			)

			continue
		}

		for _, offer := range foundOffers {
			if offer.GetDomain() != domain {
				slog.Warn(
					"merchant returned unexpected domain",
					"merchant_id", merchant.MerchantID,
					"expected", domain,
					"actual", offer.GetDomain(),
				)

				continue
			}

			offers = append(
				offers,
				offer,
			)
		}
	}

	// 10 Merchant分を一度すべて統合した後、
	// Offenro側でOffer件数上限を適用する。
	if len(offers) > s.offerLimit {
		offers = offers[:s.offerLimit]
	}

	return offers, nil
}
