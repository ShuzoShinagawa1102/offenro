package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	postgresdb "github.com/ShuzoShinagawa1102/offenro/internal/platform/postgres/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ merchant.ManagementRepository = (*Store)(nil)

func (s *Store) CreateMerchant(ctx context.Context, value model.Merchant) error {
	err := s.queries.CreateMerchant(ctx, postgresdb.CreateMerchantParams{
		MerchantID: string(value.ID), Name: value.Name, Status: string(value.Status),
	})
	return managementWriteError(err, "create merchant")
}

func (s *Store) GetMerchant(ctx context.Context, merchantID model.MerchantID) (model.Merchant, error) {
	row, err := s.queries.GetMerchant(ctx, string(merchantID))
	if err := managementReadError(err, "merchant"); err != nil {
		return model.Merchant{}, err
	}
	return model.Merchant{ID: model.MerchantID(row.MerchantID), Name: row.Name, Status: model.MerchantStatus(row.Status)}, nil
}

func (s *Store) UpdateMerchant(ctx context.Context, value model.Merchant) error {
	rows, err := s.queries.UpdateMerchant(ctx, postgresdb.UpdateMerchantParams{
		MerchantID: string(value.ID), Name: value.Name, Status: string(value.Status),
	})
	if err != nil {
		return managementWriteError(err, "update merchant")
	}
	return managementRows(rows, "merchant")
}

func (s *Store) ListCommerceDomains(ctx context.Context) ([]model.CommerceDomain, error) {
	rows, err := s.queries.ListCommerceDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("list commerce domains: %w", err)
	}
	values := make([]model.CommerceDomain, 0, len(rows))
	for _, row := range rows {
		values = append(values, model.CommerceDomain{
			ID: model.Domain(row.DomainID), Name: row.Name, Status: model.DomainStatus(row.Status),
			ProtocolVersion: row.ProtocolVersion,
		})
	}
	return values, nil
}

func (s *Store) GetCommerceDomain(ctx context.Context, domain model.Domain) (model.CommerceDomain, error) {
	row, err := s.queries.GetCommerceDomain(ctx, domain.String())
	if err := managementReadError(err, "commerce domain"); err != nil {
		return model.CommerceDomain{}, err
	}
	return model.CommerceDomain{
		ID: model.Domain(row.DomainID), Name: row.Name, Status: model.DomainStatus(row.Status),
		ProtocolVersion: row.ProtocolVersion,
	}, nil
}

func (s *Store) CreateCapability(ctx context.Context, value model.MerchantCapability) error {
	err := s.queries.CreateMerchantCapability(ctx, postgresdb.CreateMerchantCapabilityParams{
		CapabilityID: value.ID, MerchantID: string(value.Merchant.ID), DomainID: value.Domain.String(),
		ApiBaseUrl: value.APIBaseURL, Status: string(value.Status), ProtocolVersion: value.ProtocolVersion,
	})
	return managementWriteError(err, "create merchant capability")
}

func (s *Store) GetCapability(ctx context.Context, capabilityID string) (model.MerchantCapability, error) {
	row, err := s.queries.GetMerchantCapability(ctx, capabilityID)
	if err := managementReadError(err, "merchant capability"); err != nil {
		return model.MerchantCapability{}, err
	}
	return managementCapability(
		row.CapabilityID, row.MerchantID, row.MerchantName, row.MerchantStatus,
		row.DomainID, row.ApiBaseUrl, row.CapabilityStatus, row.ProtocolVersion,
	), nil
}

func (s *Store) ListCapabilities(ctx context.Context, merchantID model.MerchantID) ([]model.MerchantCapability, error) {
	rows, err := s.queries.ListMerchantCapabilities(ctx, string(merchantID))
	if err != nil {
		return nil, fmt.Errorf("list merchant capabilities: %w", err)
	}
	values := make([]model.MerchantCapability, 0, len(rows))
	for _, row := range rows {
		values = append(values, managementCapability(
			row.CapabilityID, row.MerchantID, row.MerchantName, row.MerchantStatus,
			row.DomainID, row.ApiBaseUrl, row.CapabilityStatus, row.ProtocolVersion,
		))
	}
	return values, nil
}

func (s *Store) UpdateCapability(ctx context.Context, value model.MerchantCapability) error {
	rows, err := s.queries.UpdateMerchantCapability(ctx, postgresdb.UpdateMerchantCapabilityParams{
		CapabilityID: value.ID, ApiBaseUrl: value.APIBaseURL,
		Status: string(value.Status), ProtocolVersion: value.ProtocolVersion,
	})
	if err != nil {
		return managementWriteError(err, "update merchant capability")
	}
	return managementRows(rows, "merchant capability")
}

func (s *Store) ActivateCapability(
	ctx context.Context,
	capabilityID string,
	entries []model.DiscoveryIndexEntry,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin capability activation: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := s.queries.WithTx(tx)

	rows, err := queries.SetMerchantCapabilityStatus(ctx, postgresdb.SetMerchantCapabilityStatusParams{
		CapabilityID: capabilityID, Status: string(model.CapabilityStatusActive),
	})
	if err != nil {
		return fmt.Errorf("activate merchant capability: %w", err)
	}
	if err := managementRows(rows, "merchant capability"); err != nil {
		return err
	}
	rows, err = queries.ActivateMerchantForCapability(ctx, capabilityID)
	if err != nil {
		return fmt.Errorf("activate merchant: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("%w: merchant cannot be activated", merchant.ErrConflict)
	}
	if err := queries.DeleteDiscoveryIndexByCapability(ctx, capabilityID); err != nil {
		return fmt.Errorf("clear capability discovery index: %w", err)
	}
	for _, entry := range entries {
		indexID, err := newID("index")
		if err != nil {
			return err
		}
		rows, err := queries.CreateDiscoveryIndexEntry(ctx, postgresdb.CreateDiscoveryIndexEntryParams{
			IndexEntryID: indexID, Dimension: entry.Dimension, Value: entry.Value,
			SupplyCount: int32(entry.SupplyCount), IndexedAt: timestamp(entry.IndexedAt),
			MerchantID: string(entry.MerchantID), DomainID: entry.Domain.String(),
		})
		if err != nil {
			return fmt.Errorf("create capability discovery index: %w", err)
		}
		if rows != 1 {
			return fmt.Errorf("%w: capability index does not match activation", merchant.ErrConflict)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit capability activation: %w", err)
	}
	return nil
}

func (s *Store) MarkCapabilityVerificationFailed(ctx context.Context, capabilityID string) error {
	rows, err := s.queries.SetMerchantCapabilityStatus(ctx, postgresdb.SetMerchantCapabilityStatusParams{
		CapabilityID: capabilityID, Status: string(model.CapabilityStatusVerificationFailed),
	})
	if err != nil {
		return fmt.Errorf("mark capability verification failed: %w", err)
	}
	return managementRows(rows, "merchant capability")
}

func (s *Store) CreateIncentiveRule(ctx context.Context, value model.IncentiveRule) error {
	reward, err := numeric(value.RewardValue)
	if err != nil {
		return err
	}
	err = s.queries.CreateIncentiveRule(ctx, postgresdb.CreateIncentiveRuleParams{
		IncentiveRuleID: value.ID, CapabilityID: value.CapabilityID,
		RewardType: string(value.RewardType), RewardValue: reward,
		ValidFrom: timestamp(value.ValidFrom), ValidTo: optionalTimestamp(value.ValidTo),
		Status: string(value.Status),
	})
	return managementWriteError(err, "create incentive rule")
}

func (s *Store) GetIncentiveRule(ctx context.Context, incentiveRuleID string) (model.IncentiveRule, error) {
	row, err := s.queries.GetIncentiveRule(ctx, incentiveRuleID)
	if err := managementReadError(err, "incentive rule"); err != nil {
		return model.IncentiveRule{}, err
	}
	return managementIncentive(
		row.IncentiveRuleID, row.CapabilityID, row.RewardType, row.RewardValue,
		row.ValidFrom, row.ValidTo, row.Status,
	), nil
}

func (s *Store) ListIncentiveRules(ctx context.Context, capabilityID string) ([]model.IncentiveRule, error) {
	rows, err := s.queries.ListIncentiveRules(ctx, capabilityID)
	if err != nil {
		return nil, fmt.Errorf("list incentive rules: %w", err)
	}
	values := make([]model.IncentiveRule, 0, len(rows))
	for _, row := range rows {
		values = append(values, managementIncentive(
			row.IncentiveRuleID, row.CapabilityID, row.RewardType, row.RewardValue,
			row.ValidFrom, row.ValidTo, row.Status,
		))
	}
	return values, nil
}

func (s *Store) UpdateIncentiveRule(ctx context.Context, value model.IncentiveRule) error {
	reward, err := numeric(value.RewardValue)
	if err != nil {
		return err
	}
	rows, err := s.queries.UpdateIncentiveRule(ctx, postgresdb.UpdateIncentiveRuleParams{
		IncentiveRuleID: value.ID, RewardValue: reward, ValidFrom: timestamp(value.ValidFrom),
		ValidTo: optionalTimestamp(value.ValidTo), Status: string(value.Status),
	})
	if err != nil {
		return managementWriteError(err, "update incentive rule")
	}
	return managementRows(rows, "incentive rule")
}

func (s *Store) CreateAgent(ctx context.Context, value model.Agent) error {
	err := s.queries.CreateAgent(ctx, postgresdb.CreateAgentParams{
		AgentID: value.ID, Name: value.Name, Status: string(value.Status),
	})
	return managementWriteError(err, "create agent")
}

func (s *Store) GetAgent(ctx context.Context, agentID string) (model.Agent, error) {
	row, err := s.queries.GetAgent(ctx, agentID)
	if err := managementReadError(err, "agent"); err != nil {
		return model.Agent{}, err
	}
	return model.Agent{ID: row.AgentID, Name: row.Name, Status: model.AgentStatus(row.Status)}, nil
}

func managementCapability(
	capabilityID, merchantID, merchantName, merchantStatus,
	domainID, apiBaseURL, capabilityStatus, protocolVersion string,
) model.MerchantCapability {
	return model.MerchantCapability{
		ID: capabilityID, Domain: model.Domain(domainID), APIBaseURL: apiBaseURL,
		Status: model.CapabilityStatus(capabilityStatus), ProtocolVersion: protocolVersion,
		Merchant: model.Merchant{
			ID: model.MerchantID(merchantID), Name: merchantName,
			Status: model.MerchantStatus(merchantStatus),
		},
	}
}

func managementIncentive(
	id, capabilityID, rewardType, rewardValue string,
	validFrom, validTo pgtype.Timestamptz,
	status string,
) model.IncentiveRule {
	return model.IncentiveRule{
		ID: id, CapabilityID: capabilityID, RewardType: model.RewardType(rewardType),
		RewardValue: rewardValue, ValidFrom: validFrom.Time, ValidTo: timestampPointer(validTo),
		Status: model.IncentiveStatus(status),
	}
}

func numeric(value string) (pgtype.Numeric, error) {
	var result pgtype.Numeric
	if err := result.Scan(value); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("convert numeric value: %w", err)
	}
	return result, nil
}

func managementReadError(err error, resource string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %s", merchant.ErrNotFound, resource)
	}
	if err != nil {
		return fmt.Errorf("get %s: %w", resource, err)
	}
	return nil
}

func managementWriteError(err error, operation string) error {
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return fmt.Errorf("%w: %s", merchant.ErrConflict, operation)
		case "23503", "23514":
			return fmt.Errorf("%w: %s", merchant.ErrInvalidInput, operation)
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func managementRows(rows int64, resource string) error {
	if rows != 1 {
		return fmt.Errorf("%w: %s", merchant.ErrNotFound, resource)
	}
	return nil
}
