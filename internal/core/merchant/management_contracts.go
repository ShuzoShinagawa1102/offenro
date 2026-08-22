package merchant

import (
	"context"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type ManagementRepository interface {
	CreateMerchant(ctx context.Context, value model.Merchant) error
	GetMerchant(ctx context.Context, merchantID model.MerchantID) (model.Merchant, error)
	UpdateMerchant(ctx context.Context, value model.Merchant) error

	ListCommerceDomains(ctx context.Context) ([]model.CommerceDomain, error)
	GetCommerceDomain(ctx context.Context, domain model.Domain) (model.CommerceDomain, error)

	CreateCapability(ctx context.Context, value model.MerchantCapability) error
	GetCapability(ctx context.Context, capabilityID string) (model.MerchantCapability, error)
	ListCapabilities(ctx context.Context, merchantID model.MerchantID) ([]model.MerchantCapability, error)
	UpdateCapability(ctx context.Context, value model.MerchantCapability) error
	ActivateCapability(
		ctx context.Context,
		capabilityID string,
		entries []model.DiscoveryIndexEntry,
	) error
	MarkCapabilityVerificationFailed(ctx context.Context, capabilityID string) error

	CreateIncentiveRule(ctx context.Context, value model.IncentiveRule) error
	GetIncentiveRule(ctx context.Context, incentiveRuleID string) (model.IncentiveRule, error)
	ListIncentiveRules(ctx context.Context, capabilityID string) ([]model.IncentiveRule, error)
	UpdateIncentiveRule(ctx context.Context, value model.IncentiveRule) error

	CreateAgent(ctx context.Context, value model.Agent) error
	GetAgent(ctx context.Context, agentID string) (model.Agent, error)
}

type CapabilityCatalog interface {
	BuildCapabilityIndex(
		ctx context.Context,
		capability model.MerchantCapability,
	) ([]model.DiscoveryIndexEntry, error)
}

type CreateMerchantInput struct {
	Name string
}

type UpdateMerchantInput struct {
	Name   *string
	Status *model.MerchantStatus
}

type CreateCapabilityInput struct {
	Domain          model.Domain
	APIBaseURL      string
	ProtocolVersion string
}

type UpdateCapabilityInput struct {
	APIBaseURL      *string
	ProtocolVersion *string
	Status          *model.CapabilityStatus
}

type CreateIncentiveRuleInput struct {
	RewardType  model.RewardType
	RewardValue string
	ValidFrom   *time.Time
	ValidTo     *time.Time
}

type UpdateIncentiveRuleInput struct {
	RewardValue *string
	ValidFrom   *time.Time
	ValidTo     *time.Time
	Status      *model.IncentiveStatus
}

type CreateAgentInput struct {
	Name string
}

type ManagementUseCases interface {
	CreateMerchant(ctx context.Context, input CreateMerchantInput) (model.Merchant, error)
	GetMerchant(ctx context.Context, merchantID string) (model.Merchant, error)
	UpdateMerchant(ctx context.Context, merchantID string, input UpdateMerchantInput) (model.Merchant, error)
	ListCommerceDomains(ctx context.Context) ([]model.CommerceDomain, error)
	CreateCapability(ctx context.Context, merchantID string, input CreateCapabilityInput) (model.MerchantCapability, error)
	GetCapability(ctx context.Context, capabilityID string) (model.MerchantCapability, error)
	ListCapabilities(ctx context.Context, merchantID string) ([]model.MerchantCapability, error)
	UpdateCapability(ctx context.Context, capabilityID string, input UpdateCapabilityInput) (model.MerchantCapability, error)
	VerifyCapability(ctx context.Context, capabilityID string) (model.MerchantCapability, error)
	CreateIncentiveRule(ctx context.Context, capabilityID string, input CreateIncentiveRuleInput) (model.IncentiveRule, error)
	GetIncentiveRule(ctx context.Context, incentiveRuleID string) (model.IncentiveRule, error)
	ListIncentiveRules(ctx context.Context, capabilityID string) ([]model.IncentiveRule, error)
	UpdateIncentiveRule(ctx context.Context, incentiveRuleID string, input UpdateIncentiveRuleInput) (model.IncentiveRule, error)
	CreateAgent(ctx context.Context, input CreateAgentInput) (model.Agent, error)
	GetAgent(ctx context.Context, agentID string) (model.Agent, error)
}
