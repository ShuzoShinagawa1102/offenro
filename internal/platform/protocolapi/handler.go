package protocolapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/cart"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/fulfillment"
	commonapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/common"
	commerceapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/protocol/commerce"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/protocol/model"
)

type Handler struct {
	carts         cart.UseCases
	confirmations fulfillment.UseCases
}

var _ commerceapi.StrictServerInterface = (*Handler)(nil)

func New(carts cart.UseCases, confirmations fulfillment.UseCases) *Handler {
	return &Handler{carts: carts, confirmations: confirmations}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	handler := commerceapi.NewStrictHandler(h, nil)
	commerceapi.HandlerFromMux(handler, mux)
}

func decodeSnapshot(value json.RawMessage) (modelapi.OfferSnapshot, error) {
	var snapshot modelapi.OfferSnapshot
	if err := json.Unmarshal(value, &snapshot); err != nil {
		return nil, fmt.Errorf("decode offer snapshot: %w", err)
	}
	return snapshot, nil
}

func apiError(err error) commonapi.Error {
	switch {
	case errors.Is(err, cart.ErrInvalidInput):
		return commonapi.Error{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, cart.ErrNotFound):
		return commonapi.Error{Code: "not_found", Message: err.Error()}
	case errors.Is(err, cart.ErrConflict):
		return commonapi.Error{Code: "conflict", Message: err.Error()}
	case errors.Is(err, cart.ErrRevalidation):
		return commonapi.Error{Code: "merchant_revalidation_failed", Message: "merchant offer revalidation failed"}
	case errors.Is(err, fulfillment.ErrInvalidInput):
		return commonapi.Error{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, fulfillment.ErrNotFound):
		return commonapi.Error{Code: "not_found", Message: err.Error()}
	case errors.Is(err, fulfillment.ErrConflict):
		return commonapi.Error{Code: "conflict", Message: err.Error()}
	default:
		return commonapi.Error{Code: "internal_error", Message: "internal server error"}
	}
}
