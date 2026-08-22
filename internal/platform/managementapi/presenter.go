package managementapi

import (
	"errors"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	commonapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/common"
	managementapi "github.com/ShuzoShinagawa1102/offenro/internal/generated/management/api"
	managementmodel "github.com/ShuzoShinagawa1102/offenro/internal/generated/management/model"
)

type errorKind int

const (
	errorInternal errorKind = iota
	errorBadRequest
	errorNotFound
	errorConflict
	errorVerification
)

func classifyError(err error) errorKind {
	switch {
	case errors.Is(err, merchant.ErrInvalidInput):
		return errorBadRequest
	case errors.Is(err, merchant.ErrNotFound):
		return errorNotFound
	case errors.Is(err, merchant.ErrConflict):
		return errorConflict
	case errors.Is(err, merchant.ErrVerification):
		return errorVerification
	default:
		return errorInternal
	}
}

func errorBody(kind errorKind, err error) commonapi.Error {
	code := "internal_error"
	switch kind {
	case errorBadRequest:
		code = "invalid_request"
	case errorNotFound:
		code = "not_found"
	case errorConflict:
		code = "conflict"
	case errorVerification:
		code = "merchant_verification_failed"
	}
	return commonapi.Error{Code: code, Message: err.Error()}
}

func badRequest(err error) managementapi.BadRequestJSONResponse {
	return managementapi.BadRequestJSONResponse(errorBody(errorBadRequest, err))
}

func notFound(err error) managementapi.NotFoundJSONResponse {
	return managementapi.NotFoundJSONResponse(errorBody(errorNotFound, err))
}

func conflict(err error) managementapi.ConflictJSONResponse {
	return managementapi.ConflictJSONResponse(errorBody(errorConflict, err))
}

func presentMerchant(value model.Merchant) managementmodel.Merchant {
	return managementmodel.Merchant{
		MerchantId: string(value.ID),
		Name:       value.Name,
		Status:     managementmodel.MerchantStatus(value.Status),
	}
}

func presentCapability(value model.MerchantCapability) managementmodel.MerchantCapability {
	return managementmodel.MerchantCapability{
		CapabilityId:    value.ID,
		MerchantId:      string(value.Merchant.ID),
		DomainId:        string(value.Domain),
		ApiBaseUrl:      value.APIBaseURL,
		ProtocolVersion: value.ProtocolVersion,
		Status:          managementmodel.CapabilityStatus(value.Status),
	}
}

func presentDomain(value model.CommerceDomain) managementmodel.CommerceDomain {
	return managementmodel.CommerceDomain{
		DomainId:        string(value.ID),
		Name:            value.Name,
		ProtocolVersion: value.ProtocolVersion,
		Status:          managementmodel.DomainStatus(value.Status),
	}
}

func presentIncentiveRule(value model.IncentiveRule) managementmodel.IncentiveRule {
	return managementmodel.IncentiveRule{
		IncentiveRuleId: value.ID,
		CapabilityId:    value.CapabilityID,
		RewardType:      managementmodel.RewardType(value.RewardType),
		RewardValue:     value.RewardValue,
		ValidFrom:       value.ValidFrom,
		ValidTo:         value.ValidTo,
		Status:          managementmodel.IncentiveStatus(value.Status),
	}
}

func presentAgent(value model.Agent) managementmodel.Agent {
	return managementmodel.Agent{
		AgentId: value.ID,
		Name:    value.Name,
		Status:  managementmodel.AgentStatus(value.Status),
	}
}
