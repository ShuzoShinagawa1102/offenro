package retailshoes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/fulfillment"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	merchantapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/retailshoes/generated/merchant"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/retailshoes/generated/model"
)

// Fulfiller contains only retail.shoes-specific order and delivery behavior.
type Fulfiller struct {
	clients merchantClientFactory
}

var _ fulfillment.DomainFulfiller = (*Fulfiller)(nil)

func NewFulfiller(httpClient *http.Client) *Fulfiller {
	return &Fulfiller{clients: &generatedMerchantClientFactory{httpClient: httpClient}}
}

func (f *Fulfiller) Domain() model.Domain { return Domain }

func (f *Fulfiller) ValidateDetails(raw json.RawMessage) error {
	_, err := retailShoesDetails(raw)
	return err
}

func (f *Fulfiller) Submit(
	ctx context.Context,
	capability model.MerchantCapability,
	request fulfillment.MerchantRequest,
) (fulfillment.MerchantResult, error) {
	details, err := retailShoesDetails(request.Details)
	if err != nil {
		return fulfillment.MerchantResult{}, err
	}
	client, err := f.clients.NewClient(capability)
	if err != nil {
		return fulfillment.MerchantResult{}, err
	}
	response, err := client.CreateRetailShoesOrderWithResponse(
		ctx,
		&merchantapi.CreateRetailShoesOrderParams{IdempotencyKey: request.IdempotencyKey},
		modelapi.RetailShoesOrderRequest{
			BuyerRef: request.BuyerRef, Details: details,
			MerchantOfferRef: request.MerchantOfferRef,
			PurchaseId:       request.PurchaseID, PurchaseItemId: request.PurchaseItemID,
		},
	)
	if err != nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("create retail shoes order: %w", err)
	}
	var result *modelapi.RetailShoesOrderResult
	switch response.StatusCode() {
	case http.StatusOK:
		result = response.JSON200
	case http.StatusCreated:
		result = response.JSON201
	case http.StatusConflict:
		result = response.JSON409
	}
	if result == nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant order returned status %d", response.StatusCode())
	}
	return retailShoesResult(*result, response.Body)
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
	response, err := client.GetRetailShoesOrderByIdempotencyKeyWithResponse(ctx, idempotencyKey)
	if err != nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("resolve retail shoes order: %w", err)
	}
	if response.StatusCode() == http.StatusNotFound {
		return fulfillment.MerchantResult{}, fulfillment.ErrMerchantRecordNotFound
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant order lookup returned status %d", response.StatusCode())
	}
	return retailShoesResult(*response.JSON200, response.Body)
}

func retailShoesDetails(raw json.RawMessage) (modelapi.RetailShoesDeliveryDetails, error) {
	var details modelapi.RetailShoesDeliveryDetails
	if len(raw) == 0 || json.Unmarshal(raw, &details) != nil {
		return details, fmt.Errorf("valid retail.shoes delivery details are required")
	}
	if strings.TrimSpace(details.RecipientName) == "" {
		return details, fmt.Errorf("recipient_name is required")
	}
	if len(strings.TrimSpace(details.CountryCode)) != 2 {
		return details, fmt.Errorf("country_code must contain two characters")
	}
	if strings.TrimSpace(details.PostalCode) == "" || strings.TrimSpace(details.AddressLine1) == "" {
		return details, fmt.Errorf("postal_code and address_line1 are required")
	}
	return details, nil
}

func retailShoesResult(value modelapi.RetailShoesOrderResult, snapshot []byte) (fulfillment.MerchantResult, error) {
	status := model.MerchantFulfillmentStatus(value.Status)
	if status != model.MerchantFulfillmentStatusPending && status != model.MerchantFulfillmentStatusConfirmed && status != model.MerchantFulfillmentStatusRejected {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant returned an invalid order status")
	}
	if status == model.MerchantFulfillmentStatusConfirmed && strings.TrimSpace(value.MerchantOrderRef) == "" {
		return fulfillment.MerchantResult{}, fmt.Errorf("merchant_order_ref is required for a confirmed order")
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
