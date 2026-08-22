package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/cart"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/extension"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	"github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel"
	travelhotelapi "github.com/ShuzoShinagawa1102/offenro/internal/platform/domainapi/travelhotel"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/httpclient"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/httpserver"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/managementapi"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/offertoken"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/postgres"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/protocolapi"
)

const managementToken = "integration-management-token"

func TestMerchantRegistrationSearchAndCheckout(t *testing.T) {
	databaseURL := os.Getenv("OFFENRO_INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("OFFENRO_INTEGRATION_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	store := postgres.NewTransactionalStore(tx)

	merchantOfferRef := "merchant-private-room-reference"
	expiresAt := time.Now().UTC().Add(20 * time.Minute).Truncate(time.Second)
	var revalidateCalled atomic.Bool
	merchantServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/travel/hotel/catalog":
			writeJSON(t, w, map[string]any{"hotels": []map[string]any{{
				"hotel_id": "hotel-1", "name": "Test Hotel", "prefecture_code": "14",
				"prefecture_name": "Kanagawa", "city": "Yokohama",
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/travel/hotel/search":
			writeJSON(t, w, map[string]any{"offers": []map[string]any{{
				"merchant_offer_ref": merchantOfferRef, "amount": 12000, "currency": "JPY",
				"offer_expires_at": expiresAt, "hotel": map[string]any{
					"hotel_id": "hotel-1", "name": "Test Hotel", "prefecture_code": "14",
					"prefecture_name": "Kanagawa", "city": "Yokohama",
				}, "stay": map[string]any{"check_in": "2026-09-10", "check_out": "2026-09-12"},
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/travel/hotel/offers/revalidate":
			var request struct {
				MerchantOfferRef string `json:"merchant_offer_ref"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.MerchantOfferRef != merchantOfferRef {
				http.Error(w, "invalid merchant offer reference", http.StatusBadRequest)
				return
			}
			revalidateCalled.Store(true)
			writeJSON(t, w, map[string]any{"offer": map[string]any{
				"merchant_offer_ref": merchantOfferRef, "amount": 12000, "currency": "JPY",
				"offer_expires_at": expiresAt, "hotel": map[string]any{
					"hotel_id": "hotel-1", "name": "Test Hotel", "prefecture_code": "14",
					"prefecture_name": "Kanagawa", "city": "Yokohama",
				}, "stay": map[string]any{"check_in": "2026-09-10", "check_out": "2026-09-12"},
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer merchantServer.Close()

	codec, err := offertoken.New("integration-secret-with-at-least-32-characters", 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	registry := extension.NewRegistry()
	merchantHTTPClient := httpclient.New(httpclient.Config{})
	if err := registry.Register(travelhotel.New(store, merchantHTTPClient, codec)); err != nil {
		t.Fatal(err)
	}
	managementService := merchant.NewManagementService(store, registry)
	searchService := search.NewService(store, registry, search.DefaultMerchantLimit, search.DefaultOfferLimit)
	cartService := cart.NewService(store, store, codec, store, registry)
	server := httptest.NewServer(httpserver.New(
		managementapi.New(managementService, managementToken),
		protocolapi.New(cartService),
		travelhotelapi.New(searchService),
	))
	defer server.Close()

	var createdMerchant struct {
		MerchantID string `json:"merchant_id"`
		Status     string `json:"status"`
	}
	requestJSON(t, server.Client(), http.MethodPost, server.URL+"/v1/management/merchants", managementToken,
		map[string]any{"name": "Integration Merchant"}, http.StatusCreated, &createdMerchant)
	if createdMerchant.Status != "PENDING" {
		t.Fatalf("merchant status = %s, want PENDING", createdMerchant.Status)
	}

	var capability struct {
		CapabilityID string `json:"capability_id"`
		Status       string `json:"status"`
	}
	requestJSON(t, server.Client(), http.MethodPost,
		server.URL+"/v1/management/merchants/"+createdMerchant.MerchantID+"/capabilities", managementToken,
		map[string]any{"domain_id": "travel.hotel", "api_base_url": merchantServer.URL, "protocol_version": "0.1.0"},
		http.StatusCreated, &capability)
	if capability.Status != "PENDING_VERIFICATION" {
		t.Fatalf("capability status = %s, want PENDING_VERIFICATION", capability.Status)
	}

	requestJSON(t, server.Client(), http.MethodPost,
		server.URL+"/v1/management/capabilities/"+capability.CapabilityID+"/incentive-rules", managementToken,
		map[string]any{"reward_type": "PERCENTAGE", "reward_value": "5.0000"}, http.StatusCreated, nil)
	requestJSON(t, server.Client(), http.MethodPost,
		server.URL+"/v1/management/capabilities/"+capability.CapabilityID+"/verify", managementToken,
		nil, http.StatusOK, &capability)
	if capability.Status != "ACTIVE" {
		t.Fatalf("verified capability status = %s, want ACTIVE", capability.Status)
	}

	var createdAgent struct {
		AgentID string `json:"agent_id"`
	}
	requestJSON(t, server.Client(), http.MethodPost, server.URL+"/v1/management/agents", managementToken,
		map[string]any{"name": "Integration Agent"}, http.StatusCreated, &createdAgent)

	var searchResponse struct {
		Offers []map[string]any `json:"offers"`
	}
	searchBody := requestJSON(t, server.Client(), http.MethodPost, server.URL+"/v1/travel/hotels/search", "",
		map[string]any{
			"destination": map[string]any{"prefecture_code": "14"},
			"stay":        map[string]any{"check_in": "2026-09-10", "check_out": "2026-09-12"},
			"guests":      map[string]any{"adults": 2, "rooms": 1},
		}, http.StatusOK, &searchResponse)
	if len(searchResponse.Offers) != 1 {
		t.Fatalf("offers = %d, want 1", len(searchResponse.Offers))
	}
	offerID, _ := searchResponse.Offers[0]["offer_id"].(string)
	if offerID == "" || strings.Contains(string(searchBody), merchantOfferRef) || strings.Contains(offerID, createdMerchant.MerchantID) {
		t.Fatalf("Agent response leaked or omitted offer reference: %s", searchBody)
	}

	var createdCart struct {
		CartID string `json:"cart_id"`
	}
	requestJSON(t, server.Client(), http.MethodPost, server.URL+"/v1/carts", "",
		map[string]any{"agent_id": createdAgent.AgentID, "buyer_ref": "buyer-integration"}, http.StatusCreated, &createdCart)
	requestJSON(t, server.Client(), http.MethodPost, server.URL+"/v1/carts/"+createdCart.CartID+"/items", "",
		map[string]any{
			"offer_id": offerID, "offer_snapshot": searchResponse.Offers[0], "amount": 12000, "currency": "JPY",
		}, http.StatusCreated, nil)

	var purchase struct {
		PurchaseID string `json:"purchase_id"`
		Status     string `json:"status"`
		Total      int64  `json:"total_amount"`
	}
	requestJSON(t, server.Client(), http.MethodPost, server.URL+"/v1/carts/"+createdCart.CartID+"/checkout", "",
		nil, http.StatusCreated, &purchase)
	if !revalidateCalled.Load() {
		t.Fatal("Merchant revalidate API was not called")
	}
	if purchase.Status != "CREATED" || purchase.Total != 12000 {
		t.Fatalf("purchase = %#v, want CREATED and revalidated total 12000", purchase)
	}

	var checkedOutCart struct {
		Status string `json:"status"`
	}
	requestJSON(t, server.Client(), http.MethodGet, server.URL+"/v1/carts/"+createdCart.CartID, "", nil,
		http.StatusOK, &checkedOutCart)
	if checkedOutCart.Status != "CHECKED_OUT" {
		t.Fatalf("cart status = %s, want CHECKED_OUT", checkedOutCart.Status)
	}
	requestJSON(t, server.Client(), http.MethodGet, server.URL+"/v1/purchases/"+purchase.PurchaseID, "", nil,
		http.StatusOK, &purchase)
	if purchase.Status != "CREATED" {
		t.Fatalf("persisted purchase status = %s, want CREATED", purchase.Status)
	}
}

func requestJSON(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	token string,
	body any,
	wantStatus int,
	target any,
) []byte {
	t.Helper()
	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	request, err := http.NewRequest(method, url, &requestBody)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	responseBody := new(bytes.Buffer)
	if _, err := responseBody.ReadFrom(response.Body); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("%s %s status = %d, want %d; body = %s", method, url, response.StatusCode, wantStatus, responseBody.String())
	}
	if target != nil {
		if err := json.Unmarshal(responseBody.Bytes(), target); err != nil {
			t.Fatalf("decode %s %s: %v; body = %s", method, url, err, responseBody.String())
		}
	}
	return responseBody.Bytes()
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("encode merchant response: %v", err)
	}
}

func TestManagementAPIRejectsMissingToken(t *testing.T) {
	server := httptest.NewServer(httpserver.New(managementapi.New(nil, managementToken)))
	defer server.Close()
	requestJSON(t, server.Client(), http.MethodGet, fmt.Sprintf("%s/v1/management/commerce-domains", server.URL), "", nil,
		http.StatusUnauthorized, nil)
}
