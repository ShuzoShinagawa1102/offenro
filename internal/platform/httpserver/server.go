package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// RouteMounter adds an Agent API transport to the shared HTTP server.
type RouteMounter interface {
	Mount(mux *http.ServeMux)
}

func New(mounters ...RouteMounter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	for _, mounter := range mounters {
		mounter.Mount(mux)
	}
	return mux
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{Status: "ok"}); err != nil {
		slog.Error("encode health response failed", "error", err)
	}
}
