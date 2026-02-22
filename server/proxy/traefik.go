package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anveesa/proxera/models"
)

// TraefikAdapter connects to Traefik via its REST API.
type TraefikAdapter struct {
	serverID   string
	serverName string
	apiURL     string
	apiToken   string
	httpClient *http.Client
}

func NewTraefikAdapter(serverID, serverName, apiURL, apiToken string) *TraefikAdapter {
	return &TraefikAdapter{
		serverID:   serverID,
		serverName: serverName,
		apiURL:     strings.TrimRight(apiURL, "/"),
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *TraefikAdapter) Type() string { return "traefik" }

func (a *TraefikAdapter) doGet(ctx context.Context, path string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.apiURL+path, nil)
	if err != nil {
		return nil, 0, err
	}
	if a.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiToken)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (a *TraefikAdapter) Ping(ctx context.Context) (int64, error) {
	start := time.Now()
	_, status, err := a.doGet(ctx, "/ping")
	if err != nil {
		return 0, err
	}
	if status != 200 {
		return 0, fmt.Errorf("ping returned %d", status)
	}
	return time.Since(start).Milliseconds(), nil
}

func (a *TraefikAdapter) GetMetrics(ctx context.Context) (*models.ServerMetrics, error) {
	body, _, err := a.doGet(ctx, "/api/overview")
	if err != nil {
		return nil, err
	}

	var overview struct {
		HTTP struct {
			Routers  struct{ Total int } `json:"routers"`
			Services struct{ Total int } `json:"services"`
		} `json:"http"`
	}
	json.Unmarshal(body, &overview) //nolint:errcheck

	return &models.ServerMetrics{
		ServerID:  a.serverID,
		Timestamp: time.Now(),
	}, nil
}

func (a *TraefikAdapter) GetConfig(ctx context.Context) (*models.ProxyConfig, error) {
	body, _, err := a.doGet(ctx, "/api/rawdata")
	if err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		ServerID:     a.serverID,
		ServerName:   a.serverName,
		ProxyType:    models.ProxyTraefik,
		Content:      string(body),
		Format:       "yaml",
		LastModified: time.Now().Format(time.RFC3339),
		IsValid:      true,
	}, nil
}

func (a *TraefikAdapter) PutConfig(_ context.Context, _ string) (*models.ConfigValidation, error) {
	return nil, &ErrNotSupported{Op: "PutConfig"}
}

func (a *TraefikAdapter) Reload(_ context.Context) error {
	return &ErrNotSupported{Op: "Reload"}
}

func (a *TraefikAdapter) TailLogs(_ context.Context) (io.ReadCloser, error) {
	return nil, &ErrNotSupported{Op: "TailLogs"}
}

func (a *TraefikAdapter) GetStatus(ctx context.Context) (string, error) {
	_, err := a.Ping(ctx)
	if err != nil {
		return "offline", nil
	}
	return "online", nil
}
