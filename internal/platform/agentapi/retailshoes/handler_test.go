package retailshoes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/extension"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	domain "github.com/ShuzoShinagawa1102/offenro/internal/domain/retailshoes"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/retailshoes/generated/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/httpserver"
)

func TestAgentAPIIsMounted(t *testing.T) {
	t.Parallel()

	indexRepository := discovery.NewInMemoryIndexRepository()
	domainRegistry := extension.NewRegistry()
	if err := domainRegistry.Register(
		domain.New(indexRepository, &http.Client{}),
	); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	searchService := search.NewService(
		merchant.NewInMemoryRegistry(nil),
		domainRegistry,
		search.DefaultMerchantLimit,
		search.DefaultOfferLimit,
	)
	handler := httpserver.New(New(searchService))

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/retail/shoes/search",
		strings.NewReader(`{
			"criteria":{"brand":"example-brand","size":"27.0","category":"sneakers"},
			"quantity":1
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body modelapi.RetailShoesSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Offers) != 0 {
		t.Fatalf("offer count = %d, want 0", len(body.Offers))
	}
}
