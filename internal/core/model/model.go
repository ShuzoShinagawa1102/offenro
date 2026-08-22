package model

import "time"

// Domain identifies a commerce domain supported by a Domain Extension.
type Domain string

func (d Domain) String() string {
	return string(d)
}

// MerchantID identifies a merchant registered with Offenro.
type MerchantID string

// Merchant is a merchant registered with Offenro.
type Merchant struct {
	ID     MerchantID
	Name   string
	Status MerchantStatus
}

// MerchantCapability declares that a merchant supports a domain.
type MerchantCapability struct {
	ID              string
	Merchant        Merchant
	Domain          Domain
	APIBaseURL      string
	Status          CapabilityStatus
	ProtocolVersion string
}

// DiscoveryIndexEntry is the generic representation stored by the Discovery Index.
type DiscoveryIndexEntry struct {
	Domain      Domain
	Dimension   string
	Value       string
	MerchantID  MerchantID
	SupplyCount int
	IndexedAt   time.Time
}
