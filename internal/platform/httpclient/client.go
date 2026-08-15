package httpclient

import (
	"net/http"
	"time"
)

const DefaultTimeout = 5 * time.Second

type Config struct {
	Timeout time.Duration
}

// New creates the shared HTTP client used by generated Merchant clients.
func New(config Config) *http.Client {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return &http.Client{Timeout: timeout}
}
