package cart

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
)

type Service struct {
	carts     CartStore
	purchases PurchaseReader
	offers    offer.Resolver
	merchants merchant.Registry
	verifiers offer.VerifierRegistry
}

var _ UseCases = (*Service)(nil)

func NewService(
	carts CartStore,
	purchases PurchaseReader,
	offers offer.Resolver,
	merchants merchant.Registry,
	verifiers offer.VerifierRegistry,
) *Service {
	return &Service{
		carts: carts, purchases: purchases, offers: offers,
		merchants: merchants, verifiers: verifiers,
	}
}

func newID(prefix string) (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(value[:]), nil
}
