package discovery

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ShuzoShinagawa1102/offenro/internal/search"
)

type Indexer struct {
	registry search.MerchantRegistry

	repository IndexRepository

	builders map[search.Domain]DomainIndexBuilder
}

func NewIndexer(
	registry search.MerchantRegistry,
	repository IndexRepository,
	builders map[search.Domain]DomainIndexBuilder,
) *Indexer {

	return &Indexer{
		registry:   registry,
		repository: repository,
		builders:   builders,
	}
}

func (i *Indexer) RebuildDomain(
	ctx context.Context,
	domain search.Domain,
) error {

	builder, ok :=
		i.builders[domain]
	if !ok {
		return fmt.Errorf(
			"index builder not found for domain: %s",
			domain,
		)
	}

	merchants, err :=
		i.registry.FindByDomain(
			ctx,
			domain,
		)
	if err != nil {
		return fmt.Errorf(
			"find merchants for indexing: %w",
			err,
		)
	}

	var entries []IndexEntry

	for _, merchant := range merchants {

		merchantEntries, err :=
			builder.Build(
				ctx,
				merchant,
			)

		if err != nil {
			return fmt.Errorf(
				"build index for merchant %s: %w",
				merchant.MerchantID,
				err,
			)
		}

		entries = append(
			entries,
			merchantEntries...,
		)
	}

	if err :=
		i.repository.ReplaceDomain(
			ctx,
			domain,
			entries,
		); err != nil {

		return fmt.Errorf(
			"replace discovery index: %w",
			err,
		)
	}

	slog.Info(
		"discovery index rebuilt",
		"domain", domain,
		"entries", len(entries),
	)

	return nil
}
