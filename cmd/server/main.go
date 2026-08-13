package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
	"github.com/ShuzoShinagawa1102/offenro/internal/merchant"
	"github.com/ShuzoShinagawa1102/offenro/internal/search"
	"github.com/ShuzoShinagawa1102/offenro/internal/server"
)

func main() {

	// Merchant API呼び出し用HTTP Client。
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Merchant Capability Registry。
	//
	// 現在はDB未導入のためインメモリ。
	// 将来的にはMerchant / MerchantCapability
	// テーブルから取得する。
	registry :=
		merchant.NewInMemoryRegistry(
			[]search.MerchantTarget{
				{
					MerchantID: "merchant_001",
					Domain:     search.DomainTravelHotel,
					BaseURL:    "http://localhost:8081",
				},
			},
		)

	// travel.hotel用Searcher。
	hotelSearcher :=
		merchant.NewTravelHotelSearcher(
			httpClient,
		)

	// Protocol Core。
	searchService :=
		search.NewService(
			registry,
			map[search.Domain]search.DomainSearcher{
				search.DomainTravelHotel: hotelSearcher,
			},
		)

	// Agent向けWeb API Server。
	s := server.New(
		searchService,
	)

	strictHandler :=
		api.NewStrictHandler(
			s,
			nil,
		)

	mux := http.NewServeMux()

	handler :=
		api.HandlerFromMux(
			strictHandler,
			mux,
		)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	slog.Info(
		"offenro server started",
		"addr",
		":8080",
	)

	if err := httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {

		slog.Error(
			"server failed",
			"error",
			err,
		)

		os.Exit(1)
	}
}
