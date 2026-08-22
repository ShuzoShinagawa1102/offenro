package model

import "time"

type MerchantStatus string

const (
	MerchantStatusPending   MerchantStatus = "PENDING"
	MerchantStatusActive    MerchantStatus = "ACTIVE"
	MerchantStatusSuspended MerchantStatus = "SUSPENDED"
	MerchantStatusClosed    MerchantStatus = "CLOSED"
)

type CapabilityStatus string

const (
	CapabilityStatusPendingVerification CapabilityStatus = "PENDING_VERIFICATION"
	CapabilityStatusActive              CapabilityStatus = "ACTIVE"
	CapabilityStatusVerificationFailed  CapabilityStatus = "VERIFICATION_FAILED"
	CapabilityStatusSuspended           CapabilityStatus = "SUSPENDED"
)

type DomainStatus string

const (
	DomainStatusActive     DomainStatus = "ACTIVE"
	DomainStatusInactive   DomainStatus = "INACTIVE"
	DomainStatusDeprecated DomainStatus = "DEPRECATED"
)

type AgentStatus string

const (
	AgentStatusPending   AgentStatus = "PENDING"
	AgentStatusActive    AgentStatus = "ACTIVE"
	AgentStatusSuspended AgentStatus = "SUSPENDED"
	AgentStatusClosed    AgentStatus = "CLOSED"
)

type RewardType string

const (
	RewardTypePercentage RewardType = "PERCENTAGE"
	RewardTypeFixed      RewardType = "FIXED"
)

type IncentiveStatus string

const (
	IncentiveStatusDraft    IncentiveStatus = "DRAFT"
	IncentiveStatusActive   IncentiveStatus = "ACTIVE"
	IncentiveStatusInactive IncentiveStatus = "INACTIVE"
	IncentiveStatusExpired  IncentiveStatus = "EXPIRED"
)

type CommerceDomain struct {
	ID              Domain
	Name            string
	Status          DomainStatus
	ProtocolVersion string
}

type Agent struct {
	ID     string
	Name   string
	Status AgentStatus
}

type IncentiveRule struct {
	ID           string
	CapabilityID string
	RewardType   RewardType
	RewardValue  string
	ValidFrom    time.Time
	ValidTo      *time.Time
	Status       IncentiveStatus
}
