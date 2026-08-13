package search

import (
	"context"
	"fmt"
	"log/slog"
)

type Service struct {
	registry  MerchantRegistry
	searchers map[Domain]DomainSearcher
}

func NewService(
	registry MerchantRegistry,
	searchers map[Domain]DomainSearcher,
) *Service {
	return &Service{
		registry:  registry,
		searchers: searchers,
	}
}

func (s *Service) SearchOffers(
	ctx context.Context,
	domain Domain,
	condition SearchCondition,
) ([]Offer, error) {

	if condition == nil {
		return nil, fmt.Errorf("condition is required")
	}

	// 指定DomainとConditionの型が一致しているか確認する。
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

	// Domainに対応するSearcherを取得する。
	searcher, ok := s.searchers[domain]
	if !ok {
		return nil, fmt.Errorf(
			"unsupported domain: %s",
			domain,
		)
	}

	// Merchant Capability Registryから、
	// Domainに対応するMerchantを取得する。
	merchants, err := s.registry.FindByDomain(
		ctx,
		domain,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"find merchants by domain: %w",
			err,
		)
	}

	// TODO:
	// 現在はDomainに対応するMerchantを全件検索対象としている。
	//
	// 将来的にはConditionを利用し、
	// 地域、商品・サービス特性、Merchant Coverageなどから
	// 問い合わせるMerchantを事前に絞り込む。
	//
	// Merchant数が増えた場合には、
	// 転置インデックス等を利用した
	// Discovery / Routing Indexの導入を検討する。

	var offers []Offer

	for _, merchant := range merchants {

		foundOffers, err := searcher.Search(
			ctx,
			merchant,
			condition,
		)
		if err != nil {
			// 1 Merchantが失敗しても
			// 他Merchantの検索は継続する。
			slog.Warn(
				"merchant search failed",
				"merchant_id", merchant.MerchantID,
				"domain", domain,
				"error", err,
			)

			continue
		}

		offers = append(
			offers,
			foundOffers...,
		)
	}

	return offers, nil
}
