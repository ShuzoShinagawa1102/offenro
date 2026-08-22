package offer

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
)

var (
	ErrInvalidReference = errors.New("invalid offer reference")
	ErrExpiredReference = errors.New("expired offer reference")
	ErrUnavailable      = errors.New("offer is unavailable")
)

// Reference is Offenro's server-side meaning of an opaque Agent-facing offer_id.
// MerchantOfferRef must never be exposed through an Agent API.
type Reference struct {
	CapabilityID     string
	MerchantID       model.MerchantID
	Domain           model.Domain
	MerchantOfferRef string
	ExpiresAt        time.Time
}

type IssuedReference struct {
	OfferID   string
	ExpiresAt time.Time
}

type Issuer interface {
	Issue(reference Reference) (IssuedReference, error)
}

type Resolver interface {
	Resolve(offerID string) (Reference, error)
}

type Codec interface {
	Issuer
	Resolver
}

type Verified struct {
	Snapshot  json.RawMessage
	Amount    int64
	Currency  string
	ExpiresAt *time.Time
}

// Verifier performs a live check against the Merchant API immediately before
// Purchase creation. Snapshot is the Agent-facing domain offer representation.
type Verifier interface {
	Domain() model.Domain
	Revalidate(
		ctx context.Context,
		capability model.MerchantCapability,
		offerID string,
		merchantOfferRef string,
	) (Verified, error)
}

type VerifierRegistry interface {
	OfferVerifier(domain model.Domain) (Verifier, bool)
}
