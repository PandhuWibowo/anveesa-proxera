package proxy

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anveesa/proxera/models"
)

// HAProxyAdapter connects to HAProxy via its Stats / Data Plane API.
type HAProxyAdapter struct {
	serverID   string
	serverName string
	apiURL     string
	apiToken   string
	httpClient *http.Client
}

func NewHAProxyAdapter(serverID, serverName, apiURL, apiToken string) *HAProxyAdapter {
	return &HAProxyAdapter{
		serverID:   serverID,
		serverName: serverName,
		apiURL:     strings.TrimRight(apiURL, "/"),
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *HAProxyAdapter) Type() string { return "haproxy" }

func (a *HAProxyAdapter) doGet(ctx context.Context, path string) ([]byte, int, error) {
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

func (a *HAProxyAdapter) Ping(ctx context.Context) (int64, error) {
	start := time.Now()
	_, status, err := a.doGet(ctx, "/stats;csv;norefresh")
	if err != nil {
		return 0, err
	}
	if status != 200 {
		return 0, fmt.Errorf("haproxy stats returned %d", status)
	}
	return time.Since(start).Milliseconds(), nil
}

func (a *HAProxyAdapter) GetMetrics(ctx context.Context) (*models.ServerMetrics, error) {
	body, _, err := a.doGet(ctx, "/stats;csv;norefresh")
	if err != nil {
		return nil, err
	}

	m := &models.ServerMetrics{
		ServerID:  a.serverID,
		Timestamp: time.Now(),
	}

	r := csv.NewReader(strings.NewReader(string(body)))
	r.Comment = '#'
	records, err := r.ReadAll()
	if err != nil || len(records) < 2 {
		return m, nil
	}

	headers := records[0]
	colIdx := make(map[string]int, len(headers))
	for i, h := range headers {
		colIdx[strings.TrimSpace(h)] = i
	}

	var totalConns, totalReq int64
	for _, row := range records[1:] {
		if idx, ok := colIdx["scur"]; ok && idx < len(row) {
			v, _ := strconv.ParseInt(strings.TrimSpace(row[idx]), 10, 64)
			totalConns += v
		}
		if idx, ok := colIdx["req_tot"]; ok && idx < len(row) {
			v, _ := strconv.ParseInt(strings.TrimSpace(row[idx]), 10, 64)
			totalReq += v
		}
	}

	m.ActiveConnections = int(totalConns)
	m.RequestsPerSec = float64(totalReq)
	return m, nil
}

func (a *HAProxyAdapter) GetConfig(ctx context.Context) (*models.ProxyConfig, error) {
	body, _, err := a.doGet(ctx, "/v2/services/haproxy/configuration/global")
	if err != nil {
		return &models.ProxyConfig{
			ServerID:     a.serverID,
			ServerName:   a.serverName,
			ProxyType:    models.ProxyHAProxy,
			Content:      "# HAProxy config (Data Plane API not available)",
			Format:       "haproxy",
			LastModified: time.Now().Format(time.RFC3339),
			IsValid:      true,
		}, nil
	}
	return &models.ProxyConfig{
		ServerID:     a.serverID,
		ServerName:   a.serverName,
		ProxyType:    models.ProxyHAProxy,
		Content:      string(body),
		Format:       "haproxy",
		LastModified: time.Now().Format(time.RFC3339),
		IsValid:      true,
	}, nil
}

func (a *HAProxyAdapter) PutConfig(_ context.Context, _ string) (*models.ConfigValidation, error) {
	return nil, &ErrNotSupported{Op: "PutConfig"}
}

func (a *HAProxyAdapter) Reload(_ context.Context) error {
	return &ErrNotSupported{Op: "Reload"}
}

func (a *HAProxyAdapter) TailLogs(_ context.Context) (io.ReadCloser, error) {
	return nil, &ErrNotSupported{Op: "TailLogs"}
}

func (a *HAProxyAdapter) GetStatus(ctx context.Context) (string, error) {
	_, err := a.Ping(ctx)
	if err != nil {
		return "offline", nil
	}
	return "online", nil
}
