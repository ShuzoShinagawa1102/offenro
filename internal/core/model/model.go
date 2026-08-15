package model

import "time"

// Domain identifies a commerce domain supported by a Domain Extension.
type Domain string

func (d Domain) String() string {
	return string(d)
}

// MerchantID identifies a merchant registered with Offenro.
type MerchantID string

// Merchant contains the domain-independent connection information for a merchant.
type Merchant struct {
	ID      MerchantID
	BaseURL string
}

// MerchantCapability declares that a merchant supports a domain.
type MerchantCapability struct {
	Merchant Merchant
	Domain   Domain
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
