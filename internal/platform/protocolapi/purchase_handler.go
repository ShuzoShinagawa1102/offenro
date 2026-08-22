package protocolapi

import (
	"context"
	"errors"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/cart"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	commerceapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/protocol/commerce"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/protocol/model"
)

func (h *Handler) CheckoutCart(ctx context.Context, request commerceapi.CheckoutCartRequestObject) (commerceapi.CheckoutCartResponseObject, error) {
	purchase, err := h.carts.Checkout(ctx, request.CartId)
	if err != nil {
		switch {
		case errors.Is(err, cart.ErrInvalidInput):
			return commerceapi.CheckoutCart400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrNotFound):
			return commerceapi.CheckoutCart404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrConflict):
			return commerceapi.CheckoutCart409JSONResponse{ConflictJSONResponse: commerceapi.ConflictJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrRevalidation):
			return commerceapi.CheckoutCart502JSONResponse{BadGatewayJSONResponse: commerceapi.BadGatewayJSONResponse(apiError(err))}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPIPurchase(purchase)
	if err != nil {
		return nil, err
	}
	return commerceapi.CheckoutCart201JSONResponse(response), nil
}

func (h *Handler) GetPurchase(ctx context.Context, request commerceapi.GetPurchaseRequestObject) (commerceapi.GetPurchaseResponseObject, error) {
	purchase, err := h.carts.GetPurchase(ctx, request.PurchaseId)
	if err != nil {
		if errors.Is(err, cart.ErrNotFound) {
			return commerceapi.GetPurchase404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		}
		return nil, err
	}
	response, err := toAPIPurchase(purchase)
	if err != nil {
		return nil, err
	}
	return commerceapi.GetPurchase200JSONResponse(response), nil
}

func toAPIPurchase(value model.Purchase) (modelapi.Purchase, error) {
	items := make([]modelapi.PurchaseItem, 0, len(value.Items))
	for _, item := range value.Items {
		snapshot, err := decodeSnapshot(item.OfferSnapshot)
		if err != nil {
			return modelapi.Purchase{}, err
		}
		items = append(items, modelapi.PurchaseItem{ItemId: item.ID, MerchantId: string(item.MerchantID), Domain: item.Domain.String(), OfferId: item.OfferID, OfferSnapshot: snapshot, Amount: item.Amount, Currency: item.Currency})
	}
	return modelapi.Purchase{PurchaseId: value.ID, CartId: value.CartID, AgentId: value.AgentID, BuyerRef: value.BuyerRef, Status: modelapi.PurchaseStatus(value.Status), Items: items, TotalAmount: value.TotalAmount, Currency: value.Currency, PurchasedAt: value.PurchasedAt}, nil
}
