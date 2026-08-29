package travelhotel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/fulfillment"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	merchantapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/merchant"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/model"
)

// Fulfiller contains only travel.hotel-specific reservation behavior.
type Fulfiller struct {
	clients merchantClientFactory
}

var _ fulfillment.DomainFulfiller = (*Fulfiller)(nil)

func NewFulfiller(httpClient *http.Client) *Fulfiller {
	return &Fulfiller{clients: &generatedMerchantClientFactory{httpClient: httpClient}}
}

func (f *Fulfiller) Domain() model.Domain { return Domain }

func (f *Fulfiller) ValidateDetails(raw json.RawMessage) error {
	_, err := travelHotelDetails(raw)
	return err
}

func (f *Fulfiller) Submit(
	ctx context.Context,
	capability model.MerchantCapability,
	request fulfillment.MerchantRequest,
) (fulfillment.MerchantResult, error) {
	details, err := travelHotelDetails(request.Details)
	if err != nil {
		return fulfillment.MerchantResult{}, err
	}
	client, err := f.clients.NewClient(capability)
	if err != nil {
		return fulfillment.MerchantResult{}, err
	}
	response, err := client.CreateTravelHotelReservationWithResponse(
		ctx,
		&merchantapi.CreateTravelHotelReservationParams{IdempotencyKey: request.IdempotencyKey},
		modelapi.TravelHotelReservationRequest{
			BuyerRef: request.BuyerRef, Details: details,
			MerchantOfferRef: request.MerchantOfferRef,
			PurchaseId:       request.PurchaseID, PurchaseItemId: request.PurchaseItemID,
		},
	)
	if err != nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("create travel hotel reservation: %w", err)
	}
	var result *modelapi.TravelHotelReservationResult
	switch response.StatusCode() {
	case http.StatusOK:
		result = response.JSON200
	case http.StatusCreated:
		result = response.JSON201
	case http.StatusConflict:
		result = response.JSON409
	}
	if result == nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant reservation returned status %d", response.StatusCode())
	}
	return travelHotelResult(*result, response.Body)
}

func (f *Fulfiller) Resolve(
	ctx context.Context,
	capability model.MerchantCapability,
	idempotencyKey string,
) (fulfillment.MerchantResult, error) {
	client, err := f.clients.NewClient(capability)
	if err != nil {
		return fulfillment.MerchantResult{}, err
	}
	response, err := client.GetTravelHotelReservationByIdempotencyKeyWithResponse(ctx, idempotencyKey)
	if err != nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("resolve travel hotel reservation: %w", err)
	}
	if response.StatusCode() == http.StatusNotFound {
		return fulfillment.MerchantResult{}, fulfillment.ErrMerchantRecordNotFound
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant reservation lookup returned status %d", response.StatusCode())
	}
	return travelHotelResult(*response.JSON200, response.Body)
}

func travelHotelDetails(raw json.RawMessage) (modelapi.TravelHotelReservationDetails, error) {
	var details modelapi.TravelHotelReservationDetails
	if len(raw) == 0 || json.Unmarshal(raw, &details) != nil {
		return details, fmt.Errorf("valid travel.hotel reservation details are required")
	}
	if strings.TrimSpace(details.LeadGuestName) == "" {
		return details, fmt.Errorf("lead_guest_name is required")
	}
	address := strings.TrimSpace(string(details.Email))
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Address != address {
		return details, fmt.Errorf("email must be valid")
	}
	return details, nil
}

func travelHotelResult(value modelapi.TravelHotelReservationResult, snapshot []byte) (fulfillment.MerchantResult, error) {
	status := model.MerchantFulfillmentStatus(value.Status)
	if status != model.MerchantFulfillmentStatusPending && status != model.MerchantFulfillmentStatusConfirmed && status != model.MerchantFulfillmentStatusRejected {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant returned an invalid reservation status")
	}
	if status == model.MerchantFulfillmentStatusConfirmed && strings.TrimSpace(value.MerchantOrderRef) == "" {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant_order_ref is required for a confirmed reservation")
	}
	failureCode := ""
	if value.FailureCode != nil {
		failureCode = *value.FailureCode
	}
	return fulfillment.MerchantResult{
		Status: status, MerchantOrderRef: value.MerchantOrderRef,
		FailureCode: failureCode, Snapshot: append(json.RawMessage(nil), snapshot...),
	}, nil
}
