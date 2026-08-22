package offertoken

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
)

const tokenVersion = 1

type Codec struct {
	aead cipher.AEAD
	ttl  time.Duration
	now  func() time.Time
}

var _ offer.Codec = (*Codec)(nil)

type payload struct {
	Version          int    `json:"v"`
	CapabilityID     string `json:"cap"`
	MerchantID       string `json:"merchant"`
	Domain           string `json:"domain"`
	MerchantOfferRef string `json:"ref"`
	ExpiresAt        int64  `json:"exp"`
}

func New(secret string, ttl time.Duration) (*Codec, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("offer token secret must contain at least 32 characters")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("offer token TTL must be greater than zero")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create offer token cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create offer token AEAD: %w", err)
	}
	return &Codec{aead: aead, ttl: ttl, now: time.Now}, nil
}

func (c *Codec) Issue(reference offer.Reference) (offer.IssuedReference, error) {
	if strings.TrimSpace(reference.CapabilityID) == "" ||
		strings.TrimSpace(string(reference.MerchantID)) == "" ||
		strings.TrimSpace(reference.Domain.String()) == "" ||
		strings.TrimSpace(reference.MerchantOfferRef) == "" {
		return offer.IssuedReference{}, offer.ErrInvalidReference
	}
	now := c.now().UTC()
	expiresAt := reference.ExpiresAt.UTC()
	maximumExpiresAt := now.Add(c.ttl)
	if expiresAt.IsZero() || expiresAt.After(maximumExpiresAt) {
		expiresAt = maximumExpiresAt
	}
	expiresAt = time.Unix(expiresAt.Unix(), 0).UTC()
	if !expiresAt.After(now) {
		return offer.IssuedReference{}, offer.ErrExpiredReference
	}
	plain, err := json.Marshal(payload{
		Version: tokenVersion, CapabilityID: reference.CapabilityID,
		MerchantID: string(reference.MerchantID), Domain: reference.Domain.String(),
		MerchantOfferRef: reference.MerchantOfferRef, ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		return offer.IssuedReference{}, fmt.Errorf("encode offer reference: %w", err)
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return offer.IssuedReference{}, fmt.Errorf("generate offer token nonce: %w", err)
	}
	sealed := c.aead.Seal(nonce, nonce, plain, nil)
	return offer.IssuedReference{
		OfferID: base64.RawURLEncoding.EncodeToString(sealed), ExpiresAt: expiresAt,
	}, nil
}

func (c *Codec) Resolve(offerID string) (offer.Reference, error) {
	encoded := strings.TrimSpace(offerID)
	sealed, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(sealed) <= c.aead.NonceSize() {
		return offer.Reference{}, offer.ErrInvalidReference
	}
	nonce, ciphertext := sealed[:c.aead.NonceSize()], sealed[c.aead.NonceSize():]
	plain, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return offer.Reference{}, offer.ErrInvalidReference
	}
	var value payload
	if err := json.Unmarshal(plain, &value); err != nil ||
		value.Version != tokenVersion ||
		strings.TrimSpace(value.CapabilityID) == "" ||
		strings.TrimSpace(value.MerchantID) == "" ||
		strings.TrimSpace(value.Domain) == "" ||
		strings.TrimSpace(value.MerchantOfferRef) == "" {
		return offer.Reference{}, offer.ErrInvalidReference
	}
	expiresAt := time.Unix(value.ExpiresAt, 0).UTC()
	if !expiresAt.After(c.now().UTC()) {
		return offer.Reference{}, offer.ErrExpiredReference
	}
	return offer.Reference{
		CapabilityID: value.CapabilityID, MerchantID: model.MerchantID(value.MerchantID),
		Domain: model.Domain(value.Domain), MerchantOfferRef: value.MerchantOfferRef,
		ExpiresAt: expiresAt,
	}, nil
}
