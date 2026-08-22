package merchant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

type ManagementService struct {
	repository ManagementRepository
	catalog    CapabilityCatalog
}

var _ ManagementUseCases = (*ManagementService)(nil)

func NewManagementService(repository ManagementRepository, catalog CapabilityCatalog) *ManagementService {
	return &ManagementService{repository: repository, catalog: catalog}
}

func (s *ManagementService) CreateMerchant(ctx context.Context, input CreateMerchantInput) (model.Merchant, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.Merchant{}, fmt.Errorf("%w: merchant name is required", ErrInvalidInput)
	}
	id, err := managementID("merchant")
	if err != nil {
		return model.Merchant{}, err
	}
	value := model.Merchant{ID: model.MerchantID(id), Name: name, Status: model.MerchantStatusPending}
	if err := s.repository.CreateMerchant(ctx, value); err != nil {
		return model.Merchant{}, err
	}
	return value, nil
}

func (s *ManagementService) GetMerchant(ctx context.Context, merchantID string) (model.Merchant, error) {
	if strings.TrimSpace(merchantID) == "" {
		return model.Merchant{}, fmt.Errorf("%w: merchant_id is required", ErrInvalidInput)
	}
	return s.repository.GetMerchant(ctx, model.MerchantID(merchantID))
}

func (s *ManagementService) UpdateMerchant(ctx context.Context, merchantID string, input UpdateMerchantInput) (model.Merchant, error) {
	current, err := s.GetMerchant(ctx, merchantID)
	if err != nil {
		return model.Merchant{}, err
	}
	if input.Name == nil && input.Status == nil {
		return model.Merchant{}, fmt.Errorf("%w: at least one merchant field is required", ErrInvalidInput)
	}
	if input.Name != nil {
		current.Name = strings.TrimSpace(*input.Name)
		if current.Name == "" {
			return model.Merchant{}, fmt.Errorf("%w: merchant name is required", ErrInvalidInput)
		}
	}
	if input.Status != nil {
		if *input.Status == model.MerchantStatusActive && current.Status != model.MerchantStatusActive {
			return model.Merchant{}, fmt.Errorf("%w: merchant activation requires capability verification", ErrConflict)
		}
		if !validMerchantStatus(*input.Status) {
			return model.Merchant{}, fmt.Errorf("%w: invalid merchant status", ErrInvalidInput)
		}
		current.Status = *input.Status
	}
	if err := s.repository.UpdateMerchant(ctx, current); err != nil {
		return model.Merchant{}, err
	}
	return current, nil
}

func (s *ManagementService) ListCommerceDomains(ctx context.Context) ([]model.CommerceDomain, error) {
	return s.repository.ListCommerceDomains(ctx)
}

func (s *ManagementService) CreateCapability(ctx context.Context, merchantID string, input CreateCapabilityInput) (model.MerchantCapability, error) {
	merchantValue, err := s.GetMerchant(ctx, merchantID)
	if err != nil {
		return model.MerchantCapability{}, err
	}
	if merchantValue.Status == model.MerchantStatusClosed {
		return model.MerchantCapability{}, fmt.Errorf("%w: merchant is closed", ErrConflict)
	}
	domainValue, err := s.repository.GetCommerceDomain(ctx, input.Domain)
	if err != nil {
		return model.MerchantCapability{}, err
	}
	if domainValue.Status != model.DomainStatusActive {
		return model.MerchantCapability{}, fmt.Errorf("%w: commerce domain is not active", ErrConflict)
	}
	baseURL, err := normalizeAPIBaseURL(input.APIBaseURL)
	if err != nil {
		return model.MerchantCapability{}, err
	}
	protocolVersion := strings.TrimSpace(input.ProtocolVersion)
	if protocolVersion == "" {
		return model.MerchantCapability{}, fmt.Errorf("%w: protocol_version is required", ErrInvalidInput)
	}
	id, err := managementID("capability")
	if err != nil {
		return model.MerchantCapability{}, err
	}
	value := model.MerchantCapability{
		ID: id, Merchant: merchantValue, Domain: input.Domain, APIBaseURL: baseURL,
		Status: model.CapabilityStatusPendingVerification, ProtocolVersion: protocolVersion,
	}
	if err := s.repository.CreateCapability(ctx, value); err != nil {
		return model.MerchantCapability{}, err
	}
	return value, nil
}

func (s *ManagementService) GetCapability(ctx context.Context, capabilityID string) (model.MerchantCapability, error) {
	if strings.TrimSpace(capabilityID) == "" {
		return model.MerchantCapability{}, fmt.Errorf("%w: capability_id is required", ErrInvalidInput)
	}
	return s.repository.GetCapability(ctx, capabilityID)
}

func (s *ManagementService) ListCapabilities(ctx context.Context, merchantID string) ([]model.MerchantCapability, error) {
	if _, err := s.GetMerchant(ctx, merchantID); err != nil {
		return nil, err
	}
	return s.repository.ListCapabilities(ctx, model.MerchantID(merchantID))
}

func (s *ManagementService) UpdateCapability(ctx context.Context, capabilityID string, input UpdateCapabilityInput) (model.MerchantCapability, error) {
	current, err := s.GetCapability(ctx, capabilityID)
	if err != nil {
		return model.MerchantCapability{}, err
	}
	if input.APIBaseURL == nil && input.ProtocolVersion == nil && input.Status == nil {
		return model.MerchantCapability{}, fmt.Errorf("%w: at least one capability field is required", ErrInvalidInput)
	}
	configurationChanged := false
	if input.APIBaseURL != nil {
		current.APIBaseURL, err = normalizeAPIBaseURL(*input.APIBaseURL)
		if err != nil {
			return model.MerchantCapability{}, err
		}
		configurationChanged = true
	}
	if input.ProtocolVersion != nil {
		current.ProtocolVersion = strings.TrimSpace(*input.ProtocolVersion)
		if current.ProtocolVersion == "" {
			return model.MerchantCapability{}, fmt.Errorf("%w: protocol_version is required", ErrInvalidInput)
		}
		configurationChanged = true
	}
	if configurationChanged {
		current.Status = model.CapabilityStatusPendingVerification
	}
	if input.Status != nil {
		if *input.Status == model.CapabilityStatusActive && current.Status != model.CapabilityStatusActive {
			return model.MerchantCapability{}, fmt.Errorf("%w: capability activation requires verification", ErrConflict)
		}
		if !validCapabilityStatus(*input.Status) {
			return model.MerchantCapability{}, fmt.Errorf("%w: invalid capability status", ErrInvalidInput)
		}
		current.Status = *input.Status
	}
	if err := s.repository.UpdateCapability(ctx, current); err != nil {
		return model.MerchantCapability{}, err
	}
	return current, nil
}

func (s *ManagementService) VerifyCapability(ctx context.Context, capabilityID string) (model.MerchantCapability, error) {
	capability, err := s.GetCapability(ctx, capabilityID)
	if err != nil {
		return model.MerchantCapability{}, err
	}
	if capability.Status == model.CapabilityStatusSuspended {
		return model.MerchantCapability{}, fmt.Errorf("%w: suspended capability cannot be verified", ErrConflict)
	}
	if s.catalog == nil {
		return model.MerchantCapability{}, fmt.Errorf("%w: capability catalog is unavailable", ErrVerification)
	}
	entries, err := s.catalog.BuildCapabilityIndex(ctx, capability)
	if err != nil {
		if markErr := s.repository.MarkCapabilityVerificationFailed(ctx, capability.ID); markErr != nil {
			return model.MerchantCapability{}, fmt.Errorf("%w: %v; mark failed: %v", ErrVerification, err, markErr)
		}
		return model.MerchantCapability{}, fmt.Errorf("%w: %v", ErrVerification, err)
	}
	if err := s.repository.ActivateCapability(ctx, capability.ID, entries); err != nil {
		return model.MerchantCapability{}, err
	}
	return s.repository.GetCapability(ctx, capability.ID)
}

func (s *ManagementService) CreateIncentiveRule(ctx context.Context, capabilityID string, input CreateIncentiveRuleInput) (model.IncentiveRule, error) {
	if _, err := s.GetCapability(ctx, capabilityID); err != nil {
		return model.IncentiveRule{}, err
	}
	if !validRewardType(input.RewardType) || !validRewardValue(input.RewardValue) {
		return model.IncentiveRule{}, fmt.Errorf("%w: invalid reward", ErrInvalidInput)
	}
	validFrom := time.Now().UTC()
	if input.ValidFrom != nil {
		validFrom = input.ValidFrom.UTC()
	}
	if input.ValidTo != nil && !input.ValidTo.After(validFrom) {
		return model.IncentiveRule{}, fmt.Errorf("%w: valid_to must be after valid_from", ErrInvalidInput)
	}
	id, err := managementID("incentive")
	if err != nil {
		return model.IncentiveRule{}, err
	}
	value := model.IncentiveRule{
		ID: id, CapabilityID: capabilityID, RewardType: input.RewardType,
		RewardValue: strings.TrimSpace(input.RewardValue), ValidFrom: validFrom,
		ValidTo: input.ValidTo, Status: model.IncentiveStatusActive,
	}
	if err := s.repository.CreateIncentiveRule(ctx, value); err != nil {
		return model.IncentiveRule{}, err
	}
	return value, nil
}

func (s *ManagementService) GetIncentiveRule(ctx context.Context, incentiveRuleID string) (model.IncentiveRule, error) {
	if strings.TrimSpace(incentiveRuleID) == "" {
		return model.IncentiveRule{}, fmt.Errorf("%w: incentive_rule_id is required", ErrInvalidInput)
	}
	return s.repository.GetIncentiveRule(ctx, incentiveRuleID)
}

func (s *ManagementService) ListIncentiveRules(ctx context.Context, capabilityID string) ([]model.IncentiveRule, error) {
	if _, err := s.GetCapability(ctx, capabilityID); err != nil {
		return nil, err
	}
	return s.repository.ListIncentiveRules(ctx, capabilityID)
}

func (s *ManagementService) UpdateIncentiveRule(ctx context.Context, incentiveRuleID string, input UpdateIncentiveRuleInput) (model.IncentiveRule, error) {
	current, err := s.GetIncentiveRule(ctx, incentiveRuleID)
	if err != nil {
		return model.IncentiveRule{}, err
	}
	if input.RewardValue == nil && input.ValidFrom == nil && input.ValidTo == nil && input.Status == nil {
		return model.IncentiveRule{}, fmt.Errorf("%w: at least one incentive field is required", ErrInvalidInput)
	}
	if input.RewardValue != nil {
		if !validRewardValue(*input.RewardValue) {
			return model.IncentiveRule{}, fmt.Errorf("%w: invalid reward_value", ErrInvalidInput)
		}
		current.RewardValue = strings.TrimSpace(*input.RewardValue)
	}
	if input.ValidFrom != nil {
		current.ValidFrom = input.ValidFrom.UTC()
	}
	if input.ValidTo != nil {
		validTo := input.ValidTo.UTC()
		current.ValidTo = &validTo
	}
	if current.ValidTo != nil && !current.ValidTo.After(current.ValidFrom) {
		return model.IncentiveRule{}, fmt.Errorf("%w: valid_to must be after valid_from", ErrInvalidInput)
	}
	if input.Status != nil {
		if !validIncentiveStatus(*input.Status) {
			return model.IncentiveRule{}, fmt.Errorf("%w: invalid incentive status", ErrInvalidInput)
		}
		current.Status = *input.Status
	}
	if err := s.repository.UpdateIncentiveRule(ctx, current); err != nil {
		return model.IncentiveRule{}, err
	}
	return current, nil
}

func (s *ManagementService) CreateAgent(ctx context.Context, input CreateAgentInput) (model.Agent, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.Agent{}, fmt.Errorf("%w: agent name is required", ErrInvalidInput)
	}
	id, err := managementID("agent")
	if err != nil {
		return model.Agent{}, err
	}
	value := model.Agent{ID: id, Name: name, Status: model.AgentStatusActive}
	if err := s.repository.CreateAgent(ctx, value); err != nil {
		return model.Agent{}, err
	}
	return value, nil
}

func (s *ManagementService) GetAgent(ctx context.Context, agentID string) (model.Agent, error) {
	if strings.TrimSpace(agentID) == "" {
		return model.Agent{}, fmt.Errorf("%w: agent_id is required", ErrInvalidInput)
	}
	return s.repository.GetAgent(ctx, agentID)
}

func normalizeAPIBaseURL(value string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("%w: api_base_url must be an absolute http(s) URL", ErrInvalidInput)
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func validRewardValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	parsed, ok := new(big.Rat).SetString(trimmed)
	if !ok || parsed.Sign() < 0 {
		return false
	}
	parts := strings.Split(trimmed, ".")
	return len(parts) <= 2 && (len(parts) == 1 || len(parts[1]) <= 4)
}

func validMerchantStatus(value model.MerchantStatus) bool {
	switch value {
	case model.MerchantStatusPending, model.MerchantStatusActive, model.MerchantStatusSuspended, model.MerchantStatusClosed:
		return true
	default:
		return false
	}
}

func validCapabilityStatus(value model.CapabilityStatus) bool {
	switch value {
	case model.CapabilityStatusPendingVerification, model.CapabilityStatusActive, model.CapabilityStatusVerificationFailed, model.CapabilityStatusSuspended:
		return true
	default:
		return false
	}
}

func validRewardType(value model.RewardType) bool {
	return value == model.RewardTypePercentage || value == model.RewardTypeFixed
}

func validIncentiveStatus(value model.IncentiveStatus) bool {
	switch value {
	case model.IncentiveStatusDraft, model.IncentiveStatusActive, model.IncentiveStatusInactive, model.IncentiveStatusExpired:
		return true
	default:
		return false
	}
}

func managementID(prefix string) (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate %s id: %w", prefix, err)
	}
	return prefix + "_" + hex.EncodeToString(value[:]), nil
}
