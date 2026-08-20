package travelhotel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/extension"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	domain "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/httpserver"
	"github.com/ShuzoShinagawa1102/offenro/internal/testutil"
)

func TestAgentAPIIsMounted(t *testing.T) {
	t.Parallel()

	indexRepository := testutil.NewIndexRepository()
	domainRegistry := extension.NewRegistry()
	if err := domainRegistry.Register(
		domain.New(indexRepository, &http.Client{}),
	); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	searchService := search.NewService(
		testutil.NewMerchantRegistry(nil),
		domainRegistry,
		search.DefaultMerchantLimit,
		search.DefaultOfferLimit,
	)
	handler := httpserver.New(New(searchService))

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/travel/hotels/search",
		strings.NewReader(`{
			"destination":{"prefecture_code":"14"},
			"stay":{"check_in":"2026-09-10","check_out":"2026-09-12"},
			"guests":{"adults":2,"rooms":1}
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body modelapi.TravelHotelSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Offers) != 0 {
		t.Fatalf("offer count = %d, want 0", len(body.Offers))
	}
}
