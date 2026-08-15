package discovery

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type Indexer struct {
	merchants  merchant.Registry
	repository IndexRepository
	domains    DomainRegistry
}

func NewIndexer(
	merchants merchant.Registry,
	repository IndexRepository,
	domains DomainRegistry,
) *Indexer {
	return &Indexer{
		merchants:  merchants,
		repository: repository,
		domains:    domains,
	}
}

func (i *Indexer) RebuildAll(ctx context.Context) error {
	for _, domain := range i.domains.Domains() {
		if err := i.RebuildDomain(ctx, domain); err != nil {
			return err
		}
	}

	return nil
}

func (i *Indexer) RebuildDomain(ctx context.Context, domain model.Domain) error {
	builder, ok := i.domains.IndexBuilder(domain)
	if !ok {
		return fmt.Errorf("index builder not found for domain: %s", domain)
	}

	capabilities, err := i.merchants.FindByDomain(ctx, domain)
	if err != nil {
		return fmt.Errorf("find merchants for indexing: %w", err)
	}

	entries := make([]model.DiscoveryIndexEntry, 0)
	for _, capability := range capabilities {
		merchantEntries, buildErr := builder.Build(ctx, capability)
		if buildErr != nil {
			return fmt.Errorf(
				"build index for merchant %s: %w",
				capability.Merchant.ID,
				buildErr,
			)
		}
		entries = append(entries, merchantEntries...)
	}

	if err := i.repository.ReplaceDomain(ctx, domain, entries); err != nil {
		return fmt.Errorf("replace discovery index: %w", err)
	}

	slog.Info(
		"discovery index rebuilt",
		"domain", domain,
		"entries", len(entries),
	)

	return nil
}
