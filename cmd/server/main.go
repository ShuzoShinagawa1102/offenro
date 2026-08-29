package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/cart"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/extension"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/fulfillment"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	"github.com/ShuzoShinagawa1102/offenro/internal/domain/retailshoes"
	"github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel"
	retailshoesapi "github.com/ShuzoShinagawa1102/offenro/internal/platform/domainapi/retailshoes"
	travelhotelapi "github.com/ShuzoShinagawa1102/offenro/internal/platform/domainapi/travelhotel"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/httpclient"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/httpserver"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/managementapi"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/offertoken"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/postgres"
	"github.com/ShuzoShinagawa1102/offenro/internal/platform/protocolapi"
)

const serverAddress = ":8080"

func main() {
	startupContext, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()

	pool, err := postgres.Open(startupContext, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("connect to database failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	var merchantRegistry merchant.Registry = store
	var indexRepository discovery.IndexRepository = store
	var cartStore cart.CartStore = store
	var purchaseReader cart.PurchaseReader = store
	domainRegistry := extension.NewRegistry()
	merchantHTTPClient := httpclient.New(httpclient.Config{})
	managementToken := strings.TrimSpace(os.Getenv("MANAGEMENT_API_TOKEN"))
	if managementToken == "" {
		slog.Error("MANAGEMENT_API_TOKEN is required")
		os.Exit(1)
	}
	offerCodec, err := offertoken.New(os.Getenv("OFFER_TOKEN_SECRET"), 30*time.Minute)
	if err != nil {
		slog.Error("configure offer tokens failed", "error", err)
		os.Exit(1)
	}

	// 新しいDomainを追加するときは、Domain Extensionをこの一覧へ追加する。
	domainExtensions := []extension.Extension{
		travelhotel.New(indexRepository, merchantHTTPClient, offerCodec),
		retailshoes.New(indexRepository, merchantHTTPClient, offerCodec),
	}
	for _, domainExtension := range domainExtensions {
		if err := domainRegistry.Register(domainExtension); err != nil {
			slog.Error("register domain extension failed", "error", err)
			os.Exit(1)
		}
	}

	searchService := search.NewService(
		merchantRegistry,
		domainRegistry,
		search.DefaultMerchantLimit,
		search.DefaultOfferLimit,
	)
	cartService := cart.NewService(
		cartStore,
		purchaseReader,
		offerCodec,
		merchantRegistry,
		domainRegistry,
	)
	managementService := merchant.NewManagementService(store, domainRegistry)
	fulfillmentService := fulfillment.NewService(store, merchantRegistry, domainRegistry)

	// 新しいDomainのDomain API Handlerもこの一覧へ追加する。
	routeMounters := []httpserver.RouteMounter{
		managementapi.New(managementService, managementToken),
		protocolapi.New(cartService, fulfillmentService),
		travelhotelapi.New(searchService),
		retailshoesapi.New(searchService),
	}

	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: httpserver.New(routeMounters...),
	}

	slog.Info(
		"offenro server started",
		"addr", serverAddress,
		"domains", domainRegistry.Domains(),
	)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
