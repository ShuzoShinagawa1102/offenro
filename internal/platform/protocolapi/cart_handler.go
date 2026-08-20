package protocolapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/cart"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	commerceapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/protocol/commerce"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/protocol/model"
)

func (h *Handler) CreateCart(ctx context.Context, request commerceapi.CreateCartRequestObject) (commerceapi.CreateCartResponseObject, error) {
	if request.Body == nil {
		return createCartBadRequest(fmt.Errorf("%w: request body is required", cart.ErrInvalidInput)), nil
	}
	created, err := h.carts.Create(ctx, cart.CreateInput{AgentID: request.Body.AgentId, BuyerRef: request.Body.BuyerRef})
	if err != nil {
		if errors.Is(err, cart.ErrInvalidInput) {
			return createCartBadRequest(err), nil
		}
		return nil, err
	}
	response, err := toAPICart(created)
	if err != nil {
		return nil, err
	}
	return commerceapi.CreateCart201JSONResponse(response), nil
}

func (h *Handler) GetCart(ctx context.Context, request commerceapi.GetCartRequestObject) (commerceapi.GetCartResponseObject, error) {
	value, err := h.carts.Get(ctx, request.CartId)
	if err != nil {
		if errors.Is(err, cart.ErrNotFound) {
			return commerceapi.GetCart404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		}
		return nil, err
	}
	response, err := toAPICart(value)
	if err != nil {
		return nil, err
	}
	return commerceapi.GetCart200JSONResponse(response), nil
}

func (h *Handler) UpdateCart(ctx context.Context, request commerceapi.UpdateCartRequestObject) (commerceapi.UpdateCartResponseObject, error) {
	if request.Body == nil {
		return commerceapi.UpdateCart400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(fmt.Errorf("%w: request body is required", cart.ErrInvalidInput)))}, nil
	}
	value, err := h.carts.Update(ctx, request.CartId, cart.UpdateInput{BuyerRef: request.Body.BuyerRef})
	if err != nil {
		switch {
		case errors.Is(err, cart.ErrInvalidInput):
			return commerceapi.UpdateCart400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrNotFound):
			return commerceapi.UpdateCart404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrConflict):
			return commerceapi.UpdateCart409JSONResponse{ConflictJSONResponse: commerceapi.ConflictJSONResponse(apiError(err))}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPICart(value)
	if err != nil {
		return nil, err
	}
	return commerceapi.UpdateCart200JSONResponse(response), nil
}

func (h *Handler) DeleteCart(ctx context.Context, request commerceapi.DeleteCartRequestObject) (commerceapi.DeleteCartResponseObject, error) {
	err := h.carts.Delete(ctx, request.CartId)
	if err != nil {
		switch {
		case errors.Is(err, cart.ErrNotFound):
			return commerceapi.DeleteCart404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrConflict):
			return commerceapi.DeleteCart409JSONResponse{ConflictJSONResponse: commerceapi.ConflictJSONResponse(apiError(err))}, nil
		default:
			return nil, err
		}
	}
	return commerceapi.DeleteCart204Response{}, nil
}

func (h *Handler) AddCartItem(ctx context.Context, request commerceapi.AddCartItemRequestObject) (commerceapi.AddCartItemResponseObject, error) {
	if request.Body == nil {
		return commerceapi.AddCartItem400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(fmt.Errorf("%w: request body is required", cart.ErrInvalidInput)))}, nil
	}
	snapshot, err := json.Marshal(request.Body.OfferSnapshot)
	if err != nil {
		return nil, fmt.Errorf("encode offer snapshot: %w", err)
	}
	item, err := h.carts.AddItem(ctx, request.CartId, cart.AddItemInput{
		MerchantID: model.MerchantID(request.Body.MerchantId), Domain: model.Domain(request.Body.Domain),
		OfferID: request.Body.OfferId, OfferSnapshot: snapshot, Amount: request.Body.Amount,
		Currency: request.Body.Currency, OfferExpiresAt: request.Body.OfferExpiresAt,
	})
	if err != nil {
		switch {
		case errors.Is(err, cart.ErrInvalidInput):
			return commerceapi.AddCartItem400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrNotFound):
			return commerceapi.AddCartItem404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrConflict):
			return commerceapi.AddCartItem409JSONResponse{ConflictJSONResponse: commerceapi.ConflictJSONResponse(apiError(err))}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPICartItem(item)
	if err != nil {
		return nil, err
	}
	return commerceapi.AddCartItem201JSONResponse(response), nil
}

func (h *Handler) UpdateCartItem(ctx context.Context, request commerceapi.UpdateCartItemRequestObject) (commerceapi.UpdateCartItemResponseObject, error) {
	if request.Body == nil {
		return commerceapi.UpdateCartItem400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(fmt.Errorf("%w: request body is required", cart.ErrInvalidInput)))}, nil
	}
	input := cart.UpdateItemInput{Amount: request.Body.Amount, Currency: request.Body.Currency, OfferExpiresAt: request.Body.OfferExpiresAt}
	if request.Body.OfferSnapshot != nil {
		snapshot, err := json.Marshal(*request.Body.OfferSnapshot)
		if err != nil {
			return nil, fmt.Errorf("encode offer snapshot: %w", err)
		}
		rawSnapshot := json.RawMessage(snapshot)
		input.OfferSnapshot = &rawSnapshot
	}
	item, err := h.carts.UpdateItem(ctx, request.CartId, request.ItemId, input)
	if err != nil {
		switch {
		case errors.Is(err, cart.ErrInvalidInput):
			return commerceapi.UpdateCartItem400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrNotFound):
			return commerceapi.UpdateCartItem404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrConflict):
			return commerceapi.UpdateCartItem409JSONResponse{ConflictJSONResponse: commerceapi.ConflictJSONResponse(apiError(err))}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPICartItem(item)
	if err != nil {
		return nil, err
	}
	return commerceapi.UpdateCartItem200JSONResponse(response), nil
}

func (h *Handler) DeleteCartItem(ctx context.Context, request commerceapi.DeleteCartItemRequestObject) (commerceapi.DeleteCartItemResponseObject, error) {
	err := h.carts.DeleteItem(ctx, request.CartId, request.ItemId)
	if err != nil {
		switch {
		case errors.Is(err, cart.ErrNotFound):
			return commerceapi.DeleteCartItem404JSONResponse{NotFoundJSONResponse: commerceapi.NotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, cart.ErrConflict):
			return commerceapi.DeleteCartItem409JSONResponse{ConflictJSONResponse: commerceapi.ConflictJSONResponse(apiError(err))}, nil
		default:
			return nil, err
		}
	}
	return commerceapi.DeleteCartItem204Response{}, nil
}

func toAPICart(value model.Cart) (modelapi.Cart, error) {
	items := make([]modelapi.CartItem, 0, len(value.Items))
	for _, item := range value.Items {
		apiItem, err := toAPICartItem(item)
		if err != nil {
			return modelapi.Cart{}, err
		}
		items = append(items, apiItem)
	}
	return modelapi.Cart{CartId: value.ID, AgentId: value.AgentID, BuyerRef: value.BuyerRef, Status: modelapi.CartStatus(value.Status), Items: items, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}, nil
}

func toAPICartItem(value model.CartItem) (modelapi.CartItem, error) {
	snapshot, err := decodeSnapshot(value.OfferSnapshot)
	if err != nil {
		return modelapi.CartItem{}, err
	}
	return modelapi.CartItem{ItemId: value.ID, MerchantId: string(value.MerchantID), Domain: value.Domain.String(), OfferId: value.OfferID, OfferSnapshot: snapshot, Amount: value.Amount, Currency: value.Currency, OfferExpiresAt: value.OfferExpiresAt, Status: modelapi.CartItemStatus(value.Status)}, nil
}

func createCartBadRequest(err error) commerceapi.CreateCart400JSONResponse {
	return commerceapi.CreateCart400JSONResponse{BadRequestJSONResponse: commerceapi.BadRequestJSONResponse(apiError(err))}
}
