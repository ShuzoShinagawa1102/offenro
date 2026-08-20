package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	postgresdb "github.com/ShuzoShinagawa1102/offenro/internal/platform/postgres/generated"
)

var _ discovery.IndexRepository = (*Store)(nil)

func (s *Store) ReplaceDomain(ctx context.Context, domain model.Domain, entries []model.DiscoveryIndexEntry) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin discovery index transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := s.queries.WithTx(tx)

	if err := queries.DeleteDiscoveryIndexByDomain(ctx, domain.String()); err != nil {
		return fmt.Errorf("delete discovery index: %w", err)
	}

	for _, entry := range entries {
		if entry.Domain != domain {
			return fmt.Errorf("discovery index domain mismatch: %s", entry.Domain)
		}
		if entry.SupplyCount < 0 || entry.SupplyCount > math.MaxInt32 {
			return fmt.Errorf("invalid discovery supply count: %d", entry.SupplyCount)
		}
		indexEntryID, err := newID("index")
		if err != nil {
			return err
		}

		rowsAffected, err := queries.CreateDiscoveryIndexEntry(ctx, postgresdb.CreateDiscoveryIndexEntryParams{
			IndexEntryID: indexEntryID,
			Dimension:    entry.Dimension,
			Value:        entry.Value,
			SupplyCount:  int32(entry.SupplyCount),
			IndexedAt:    timestamp(entry.IndexedAt),
			MerchantID:   string(entry.MerchantID),
			DomainID:     domain.String(),
		})
		if err != nil {
			return fmt.Errorf("create discovery index entry: %w", err)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("active merchant capability not found: merchant=%s domain=%s", entry.MerchantID, domain)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit discovery index: %w", err)
	}
	return nil
}

func (s *Store) Find(
	ctx context.Context,
	domain model.Domain,
	dimension string,
	value string,
	limit int,
) ([]model.DiscoveryIndexEntry, error) {
	resultLimit := int32(math.MaxInt32)
	if limit > 0 && limit < math.MaxInt32 {
		resultLimit = int32(limit)
	}
	rows, err := s.queries.FindDiscoveryIndexEntries(ctx, postgresdb.FindDiscoveryIndexEntriesParams{
		DomainID: domain.String(), Dimension: dimension, Value: value, ResultLimit: resultLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("find discovery index entries: %w", err)
	}

	entries := make([]model.DiscoveryIndexEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, model.DiscoveryIndexEntry{
			Domain: model.Domain(row.DomainID), Dimension: row.Dimension, Value: row.Value,
			MerchantID: model.MerchantID(row.MerchantID), SupplyCount: int(row.SupplyCount),
			IndexedAt: row.IndexedAt.Time,
		})
	}
	return entries, nil
}
