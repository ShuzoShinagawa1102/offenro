package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
	"github.com/ShuzoShinagawa1102/offenro/internal/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/mockdata"
	"github.com/ShuzoShinagawa1102/offenro/internal/search"
	"github.com/ShuzoShinagawa1102/offenro/internal/server"
)

const (
	merchantSearchLimit = 10
	offerSearchLimit    = 50
)

func main() {

	httpClient :=
		&http.Client{
			Timeout:
				5 * time.Second,
		}

	// ---------------------------------
	// Merchant Capability Registry
	// ---------------------------------

	targets :=
		make(
			[]search.MerchantTarget,
			0,
			30,
		)

	for _, master :=
		range mockdata.MerchantMasters() {

		targets = append(
			targets,
			search.MerchantTarget{
				MerchantID:
					master.ID,

				Domain:
					search.DomainTravelHotel,

				BaseURL:
					fmt.Sprintf(
						"http://localhost:8081/merchants/%s",
						master.ID,
					),
			},
		)
	}

	registry :=
		merchant.
			NewInMemoryCapabilityRegistry(
				targets,
			)

	// ---------------------------------
	// Discovery Index
	// ---------------------------------

	indexRepository :=
		discovery.
			NewInMemoryIndexRepository()

	catalogClient :=
		merchant.
			NewTravelHotelCatalogClient(
				httpClient,
			)

	hotelIndexBuilder :=
		discovery.
			NewTravelHotelIndexBuilder(
				catalogClient,
			)

	indexer :=
		discovery.NewIndexer(
			registry,
			indexRepository,
			map[search.Domain]discovery.DomainIndexBuilder{
				search.DomainTravelHotel:
					hotelIndexBuilder,
			},
		)

	// 現在は永続Indexを持たないので、
	// Server起動時にMerchant Catalogから再構築する。
	//
	// 将来的には別プロセス
	// cmd/discovery-indexer等へ切り出す。
	if err :=
		indexer.RebuildDomain(
			context.Background(),
			search.DomainTravelHotel,
		); err != nil {

		slog.Error(
			"failed to build discovery index",
			"error", err,
		)

		os.Exit(1)
	}

	// ---------------------------------
	// Merchant Discovery
	// ---------------------------------

	hotelDiscovery :=
		discovery.
			NewTravelHotelMerchantDiscovery(
				indexRepository,
			)

	// ---------------------------------
	// Live Search
	// ---------------------------------

	hotelSearcher :=
		merchant.
			NewTravelHotelSearcher(
				httpClient,
			)

	searchService :=
		search.NewService(
			registry,

			map[search.Domain]search.MerchantDiscovery{
				search.DomainTravelHotel:
					hotelDiscovery,
			},

			map[search.Domain]search.DomainSearcher{
				search.DomainTravelHotel:
					hotelSearcher,
			},

			merchantSearchLimit,
			offerSearchLimit,
		)

	// ---------------------------------
	// Agent Web API
	// ---------------------------------

	s :=
		server.New(
			searchService,
		)

	strictHandler :=
		api.NewStrictHandler(
			s,
			nil,
		)

	mux :=
		http.NewServeMux()

	handler :=
		api.HandlerFromMux(
			strictHandler,
			mux,
		)

	httpServer :=
		&http.Server{
			Addr:    ":8080",
			Handler: handler,
		}

	slog.Info(
		"offenro server started",
		"addr", ":8080",
		"merchant_limit",
		merchantSearchLimit,
		"offer_limit",
		offerSearchLimit,
	)

	if err :=
		httpServer.ListenAndServe();
		err != nil &&
			err != http.ErrServerClosed {

		slog.Error(
			"server failed",
			"error", err,
		)

		os.Exit(1)
	}
}
