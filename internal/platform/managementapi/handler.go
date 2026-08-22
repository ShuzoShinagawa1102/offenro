package managementapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	commonapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/common"
	managementapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/management/api"
)

type Handler struct {
	service merchant.ManagementUseCases
	token   string
}

var _ managementapi.StrictServerInterface = (*Handler)(nil)

func New(service merchant.ManagementUseCases, token string) *Handler {
	return &Handler{service: service, token: strings.TrimSpace(token)}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	managementMux := http.NewServeMux()
	strictHandler := managementapi.NewStrictHandler(h, nil)
	managementapi.HandlerFromMux(strictHandler, managementMux)
	mux.Handle("/v1/management/", h.requireBearer(managementMux))
}

func (h *Handler) requireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if h.token == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(h.token)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(commonapi.Error{Code: "unauthorized", Message: "valid management bearer token is required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func missingBody() error {
	return errors.New("request body is required")
}

func (h *Handler) CreateMerchant(ctx context.Context, request managementapi.CreateMerchantRequestObject) (managementapi.CreateMerchantResponseObject, error) {
	if request.Body == nil {
		return managementapi.CreateMerchant400JSONResponse{BadRequestJSONResponse: badRequest(missingBody())}, nil
	}
	value, err := h.service.CreateMerchant(ctx, merchant.CreateMerchantInput{Name: request.Body.Name})
	if err != nil {
		if classifyError(err) == errorBadRequest {
			return managementapi.CreateMerchant400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
		}
		return nil, err
	}
	return managementapi.CreateMerchant201JSONResponse(presentMerchant(value)), nil
}

func (h *Handler) GetMerchant(ctx context.Context, request managementapi.GetMerchantRequestObject) (managementapi.GetMerchantResponseObject, error) {
	value, err := h.service.GetMerchant(ctx, string(request.MerchantId))
	if err != nil {
		if classifyError(err) == errorNotFound {
			return managementapi.GetMerchant404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		}
		return nil, err
	}
	return managementapi.GetMerchant200JSONResponse(presentMerchant(value)), nil
}

func (h *Handler) UpdateMerchant(ctx context.Context, request managementapi.UpdateMerchantRequestObject) (managementapi.UpdateMerchantResponseObject, error) {
	if request.Body == nil {
		return managementapi.UpdateMerchant400JSONResponse{BadRequestJSONResponse: badRequest(missingBody())}, nil
	}
	var status *model.MerchantStatus
	if request.Body.Status != nil {
		value := model.MerchantStatus(*request.Body.Status)
		status = &value
	}
	value, err := h.service.UpdateMerchant(ctx, string(request.MerchantId), merchant.UpdateMerchantInput{
		Name: request.Body.Name, Status: status,
	})
	if err != nil {
		switch classifyError(err) {
		case errorBadRequest:
			return managementapi.UpdateMerchant400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
		case errorNotFound:
			return managementapi.UpdateMerchant404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		case errorConflict:
			return managementapi.UpdateMerchant409JSONResponse{ConflictJSONResponse: conflict(err)}, nil
		default:
			return nil, err
		}
	}
	return managementapi.UpdateMerchant200JSONResponse(presentMerchant(value)), nil
}

func (h *Handler) ListCommerceDomains(ctx context.Context, _ managementapi.ListCommerceDomainsRequestObject) (managementapi.ListCommerceDomainsResponseObject, error) {
	values, err := h.service.ListCommerceDomains(ctx)
	if err != nil {
		return nil, err
	}
	response := make(managementapi.ListCommerceDomains200JSONResponse, 0, len(values))
	for _, value := range values {
		response = append(response, presentDomain(value))
	}
	return response, nil
}

func (h *Handler) CreateMerchantCapability(ctx context.Context, request managementapi.CreateMerchantCapabilityRequestObject) (managementapi.CreateMerchantCapabilityResponseObject, error) {
	if request.Body == nil {
		return managementapi.CreateMerchantCapability400JSONResponse{BadRequestJSONResponse: badRequest(missingBody())}, nil
	}
	value, err := h.service.CreateCapability(ctx, string(request.MerchantId), merchant.CreateCapabilityInput{
		Domain: model.Domain(request.Body.DomainId), APIBaseURL: request.Body.ApiBaseUrl, ProtocolVersion: request.Body.ProtocolVersion,
	})
	if err != nil {
		switch classifyError(err) {
		case errorBadRequest:
			return managementapi.CreateMerchantCapability400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
		case errorNotFound:
			return managementapi.CreateMerchantCapability404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		case errorConflict:
			return managementapi.CreateMerchantCapability409JSONResponse{ConflictJSONResponse: conflict(err)}, nil
		default:
			return nil, err
		}
	}
	return managementapi.CreateMerchantCapability201JSONResponse(presentCapability(value)), nil
}

func (h *Handler) GetMerchantCapability(ctx context.Context, request managementapi.GetMerchantCapabilityRequestObject) (managementapi.GetMerchantCapabilityResponseObject, error) {
	value, err := h.service.GetCapability(ctx, string(request.CapabilityId))
	if err != nil {
		if classifyError(err) == errorNotFound {
			return managementapi.GetMerchantCapability404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		}
		return nil, err
	}
	return managementapi.GetMerchantCapability200JSONResponse(presentCapability(value)), nil
}

func (h *Handler) ListMerchantCapabilities(ctx context.Context, request managementapi.ListMerchantCapabilitiesRequestObject) (managementapi.ListMerchantCapabilitiesResponseObject, error) {
	values, err := h.service.ListCapabilities(ctx, string(request.MerchantId))
	if err != nil {
		if classifyError(err) == errorNotFound {
			return managementapi.ListMerchantCapabilities404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		}
		return nil, err
	}
	response := make(managementapi.ListMerchantCapabilities200JSONResponse, 0, len(values))
	for _, value := range values {
		response = append(response, presentCapability(value))
	}
	return response, nil
}

func (h *Handler) UpdateMerchantCapability(ctx context.Context, request managementapi.UpdateMerchantCapabilityRequestObject) (managementapi.UpdateMerchantCapabilityResponseObject, error) {
	if request.Body == nil {
		return managementapi.UpdateMerchantCapability400JSONResponse{BadRequestJSONResponse: badRequest(missingBody())}, nil
	}
	var status *model.CapabilityStatus
	if request.Body.Status != nil {
		value := model.CapabilityStatus(*request.Body.Status)
		status = &value
	}
	value, err := h.service.UpdateCapability(ctx, string(request.CapabilityId), merchant.UpdateCapabilityInput{
		APIBaseURL: request.Body.ApiBaseUrl, ProtocolVersion: request.Body.ProtocolVersion, Status: status,
	})
	if err != nil {
		switch classifyError(err) {
		case errorBadRequest:
			return managementapi.UpdateMerchantCapability400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
		case errorNotFound:
			return managementapi.UpdateMerchantCapability404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		case errorConflict:
			return managementapi.UpdateMerchantCapability409JSONResponse{ConflictJSONResponse: conflict(err)}, nil
		default:
			return nil, err
		}
	}
	return managementapi.UpdateMerchantCapability200JSONResponse(presentCapability(value)), nil
}

func (h *Handler) VerifyMerchantCapability(ctx context.Context, request managementapi.VerifyMerchantCapabilityRequestObject) (managementapi.VerifyMerchantCapabilityResponseObject, error) {
	value, err := h.service.VerifyCapability(ctx, string(request.CapabilityId))
	if err != nil {
		switch classifyError(err) {
		case errorNotFound:
			return managementapi.VerifyMerchantCapability404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		case errorConflict:
			return managementapi.VerifyMerchantCapability409JSONResponse{ConflictJSONResponse: conflict(err)}, nil
		case errorVerification:
			return managementapi.VerifyMerchantCapability502JSONResponse(errorBody(errorVerification, err)), nil
		default:
			return nil, err
		}
	}
	return managementapi.VerifyMerchantCapability200JSONResponse(presentCapability(value)), nil
}

func (h *Handler) CreateIncentiveRule(ctx context.Context, request managementapi.CreateIncentiveRuleRequestObject) (managementapi.CreateIncentiveRuleResponseObject, error) {
	if request.Body == nil {
		return managementapi.CreateIncentiveRule400JSONResponse{BadRequestJSONResponse: badRequest(missingBody())}, nil
	}
	value, err := h.service.CreateIncentiveRule(ctx, string(request.CapabilityId), merchant.CreateIncentiveRuleInput{
		RewardType: model.RewardType(request.Body.RewardType), RewardValue: request.Body.RewardValue,
		ValidFrom: request.Body.ValidFrom, ValidTo: request.Body.ValidTo,
	})
	if err != nil {
		switch classifyError(err) {
		case errorBadRequest:
			return managementapi.CreateIncentiveRule400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
		case errorNotFound:
			return managementapi.CreateIncentiveRule404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		default:
			return nil, err
		}
	}
	return managementapi.CreateIncentiveRule201JSONResponse(presentIncentiveRule(value)), nil
}

func (h *Handler) GetIncentiveRule(ctx context.Context, request managementapi.GetIncentiveRuleRequestObject) (managementapi.GetIncentiveRuleResponseObject, error) {
	value, err := h.service.GetIncentiveRule(ctx, string(request.IncentiveRuleId))
	if err != nil {
		if classifyError(err) == errorNotFound {
			return managementapi.GetIncentiveRule404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		}
		return nil, err
	}
	return managementapi.GetIncentiveRule200JSONResponse(presentIncentiveRule(value)), nil
}

func (h *Handler) ListIncentiveRules(ctx context.Context, request managementapi.ListIncentiveRulesRequestObject) (managementapi.ListIncentiveRulesResponseObject, error) {
	values, err := h.service.ListIncentiveRules(ctx, string(request.CapabilityId))
	if err != nil {
		if classifyError(err) == errorNotFound {
			return managementapi.ListIncentiveRules404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		}
		return nil, err
	}
	response := make(managementapi.ListIncentiveRules200JSONResponse, 0, len(values))
	for _, value := range values {
		response = append(response, presentIncentiveRule(value))
	}
	return response, nil
}

func (h *Handler) UpdateIncentiveRule(ctx context.Context, request managementapi.UpdateIncentiveRuleRequestObject) (managementapi.UpdateIncentiveRuleResponseObject, error) {
	if request.Body == nil {
		return managementapi.UpdateIncentiveRule400JSONResponse{BadRequestJSONResponse: badRequest(missingBody())}, nil
	}
	var status *model.IncentiveStatus
	if request.Body.Status != nil {
		value := model.IncentiveStatus(*request.Body.Status)
		status = &value
	}
	value, err := h.service.UpdateIncentiveRule(ctx, string(request.IncentiveRuleId), merchant.UpdateIncentiveRuleInput{
		RewardValue: request.Body.RewardValue, ValidFrom: request.Body.ValidFrom, ValidTo: request.Body.ValidTo, Status: status,
	})
	if err != nil {
		switch classifyError(err) {
		case errorBadRequest:
			return managementapi.UpdateIncentiveRule400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
		case errorNotFound:
			return managementapi.UpdateIncentiveRule404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		default:
			return nil, err
		}
	}
	return managementapi.UpdateIncentiveRule200JSONResponse(presentIncentiveRule(value)), nil
}

func (h *Handler) CreateAgent(ctx context.Context, request managementapi.CreateAgentRequestObject) (managementapi.CreateAgentResponseObject, error) {
	if request.Body == nil {
		return managementapi.CreateAgent400JSONResponse{BadRequestJSONResponse: badRequest(missingBody())}, nil
	}
	value, err := h.service.CreateAgent(ctx, merchant.CreateAgentInput{Name: request.Body.Name})
	if err != nil {
		if classifyError(err) == errorBadRequest {
			return managementapi.CreateAgent400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
		}
		return nil, err
	}
	return managementapi.CreateAgent201JSONResponse(presentAgent(value)), nil
}

func (h *Handler) GetAgent(ctx context.Context, request managementapi.GetAgentRequestObject) (managementapi.GetAgentResponseObject, error) {
	value, err := h.service.GetAgent(ctx, string(request.AgentId))
	if err != nil {
		if classifyError(err) == errorNotFound {
			return managementapi.GetAgent404JSONResponse{NotFoundJSONResponse: notFound(err)}, nil
		}
		return nil, err
	}
	return managementapi.GetAgent200JSONResponse(presentAgent(value)), nil
}
