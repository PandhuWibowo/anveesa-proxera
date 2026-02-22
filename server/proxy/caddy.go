package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anveesa/proxera/models"
)

// CaddyAdapter connects to Caddy via its Admin API.
type CaddyAdapter struct {
	serverID   string
	serverName string
	apiURL     string
	httpClient *http.Client
}

func NewCaddyAdapter(serverID, serverName, apiURL string) *CaddyAdapter {
	return &CaddyAdapter{
		serverID:   serverID,
		serverName: serverName,
		apiURL:     strings.TrimRight(apiURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *CaddyAdapter) Type() string { return "caddy" }

func (a *CaddyAdapter) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, a.apiURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func (a *CaddyAdapter) Ping(ctx context.Context) (int64, error) {
	start := time.Now()
	_, status, err := a.doRequest(ctx, "GET", "/config/", nil)
	if err != nil {
		return 0, err
	}
	if status != 200 {
		return 0, fmt.Errorf("caddy admin returned %d", status)
	}
	return time.Since(start).Milliseconds(), nil
}

func (a *CaddyAdapter) GetMetrics(ctx context.Context) (*models.ServerMetrics, error) {
	return &models.ServerMetrics{
		ServerID:  a.serverID,
		Timestamp: time.Now(),
	}, nil
}

func (a *CaddyAdapter) GetConfig(ctx context.Context) (*models.ProxyConfig, error) {
	body, _, err := a.doRequest(ctx, "GET", "/config/", nil)
	if err != nil {
		return nil, err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err == nil {
		body = pretty.Bytes()
	}

	return &models.ProxyConfig{
		ServerID:     a.serverID,
		ServerName:   a.serverName,
		ProxyType:    models.ProxyCaddy,
		Content:      string(body),
		Format:       "yaml",
		LastModified: time.Now().Format(time.RFC3339),
		IsValid:      true,
	}, nil
}

func (a *CaddyAdapter) PutConfig(ctx context.Context, content string) (*models.ConfigValidation, error) {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return &models.ConfigValidation{
			IsValid: false,
			Errors:  []string{"Invalid JSON: " + err.Error()},
		}, nil
	}

	_, status, err := a.doRequest(ctx, "POST", "/load", []byte(content))
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return &models.ConfigValidation{
			IsValid: false,
			Errors:  []string{fmt.Sprintf("Caddy returned HTTP %d", status)},
		}, nil
	}
	return &models.ConfigValidation{IsValid: true}, nil
}

func (a *CaddyAdapter) Reload(ctx context.Context) error {
	_, status, err := a.doRequest(ctx, "POST", "/load", nil)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("reload returned %d", status)
	}
	return nil
}

func (a *CaddyAdapter) TailLogs(_ context.Context) (io.ReadCloser, error) {
	return nil, &ErrNotSupported{Op: "TailLogs"}
}

func (a *CaddyAdapter) GetStatus(ctx context.Context) (string, error) {
	_, err := a.Ping(ctx)
	if err != nil {
		return "offline", nil
	}
	return "online", nil
}
