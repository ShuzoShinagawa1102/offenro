package offertoken

import (
	"errors"
	"testing"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
)

func TestCodecRoundTripAndTamperDetection(t *testing.T) {
	codec, err := New("0123456789abcdef0123456789abcdef", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	codec.now = func() time.Time { return now }

	issued, err := codec.Issue(offer.Reference{
		CapabilityID: "capability_1", MerchantID: model.MerchantID("merchant_1"),
		Domain: model.Domain("travel.hotel"), MerchantOfferRef: "merchant-secret-ref",
	})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := codec.Resolve(issued.OfferID)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.MerchantOfferRef != "merchant-secret-ref" || resolved.CapabilityID != "capability_1" {
		t.Fatalf("unexpected resolved reference: %#v", resolved)
	}
	if resolved.ExpiresAt != issued.ExpiresAt {
		t.Fatalf("expiry mismatch: got %s, want %s", resolved.ExpiresAt, issued.ExpiresAt)
	}

	replacement := "A"
	if issued.OfferID[0] == 'A' {
		replacement = "B"
	}
	tampered := replacement + issued.OfferID[1:]
	if _, err := codec.Resolve(tampered); !errors.Is(err, offer.ErrInvalidReference) {
		t.Fatalf("tampered token error = %v", err)
	}
}

func TestCodecRejectsExpiredReference(t *testing.T) {
	codec, err := New("0123456789abcdef0123456789abcdef", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	codec.now = func() time.Time { return now }
	_, err = codec.Issue(offer.Reference{
		CapabilityID: "capability_1", MerchantID: model.MerchantID("merchant_1"),
		Domain: model.Domain("travel.hotel"), MerchantOfferRef: "ref", ExpiresAt: now,
	})
	if !errors.Is(err, offer.ErrExpiredReference) {
		t.Fatalf("error = %v", err)
	}
}
