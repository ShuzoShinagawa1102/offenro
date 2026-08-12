package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
	"github.com/ShuzoShinagawa1102/offenro/internal/server"
)

func main() {
	s := server.New()

	strictHundler := api.NewStrictHandler(s, nil)

	mux := http.NewServeMux()

	handler := api.HandlerFromMux(strictHundler, mux)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	slog.Info("server started", "addr", ":8080")

	if err := httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
