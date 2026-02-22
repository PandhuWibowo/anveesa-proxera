package proxy

import (
	"context"
	"io"

	"github.com/anveesa/proxera/models"
)

// ProxyAdapter defines the interface every proxy backend must implement.
type ProxyAdapter interface {
	// Type returns the proxy type string (nginx, traefik, caddy, haproxy).
	Type() string
	// Ping checks reachability and returns latency in milliseconds.
	Ping(ctx context.Context) (latencyMs int64, err error)
	// GetMetrics retrieves live metrics from the proxy.
	GetMetrics(ctx context.Context) (*models.ServerMetrics, error)
	// GetConfig fetches the current proxy configuration text.
	GetConfig(ctx context.Context) (*models.ProxyConfig, error)
	// PutConfig writes and validates a new configuration.
	PutConfig(ctx context.Context, content string) (*models.ConfigValidation, error)
	// Reload triggers a graceful configuration reload.
	Reload(ctx context.Context) error
	// TailLogs returns a ReadCloser for streaming log lines.
	TailLogs(ctx context.Context) (io.ReadCloser, error)
	// GetStatus returns the current operational status string.
	GetStatus(ctx context.Context) (string, error)
}

// ErrNotSupported is returned when an operation is not supported by the adapter.
type ErrNotSupported struct {
	Op string
}

func (e *ErrNotSupported) Error() string {
	return "operation not supported by this adapter: " + e.Op
}
