package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/anveesa/proxera/models"
	"golang.org/x/crypto/ssh"
)

// NGINXAdapter connects to NGINX via SSH.
type NGINXAdapter struct {
	serverID   string
	serverName string
	host       string
	port       int
	sshUser    string
	sshKey     string // decrypted PEM private key
	sshPool    *SSHPool
}

func NewNGINXAdapter(serverID, serverName, host string, port int, sshUser, sshKey string, pool *SSHPool) *NGINXAdapter {
	return &NGINXAdapter{
		serverID:   serverID,
		serverName: serverName,
		host:       host,
		port:       port,
		sshUser:    sshUser,
		sshKey:     sshKey,
		sshPool:    pool,
	}
}

func (a *NGINXAdapter) Type() string { return "nginx" }

func (a *NGINXAdapter) getClient(ctx context.Context) (*ssh.Client, error) {
	return a.sshPool.Get(ctx, a.serverID, a.host, a.port, a.sshUser, a.sshKey)
}

func (a *NGINXAdapter) Ping(ctx context.Context) (int64, error) {
	start := time.Now()
	addr := fmt.Sprintf("%s:%d", a.host, a.port)
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start).Milliseconds(), nil
}

func (a *NGINXAdapter) GetMetrics(ctx context.Context) (*models.ServerMetrics, error) {
	client, err := a.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("ssh connect: %w", err)
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	out, err := runSession(session, "curl -sf http://127.0.0.1/nginx_status 2>/dev/null || curl -sf http://127.0.0.1:8080/nginx_status 2>/dev/null || echo 'unavailable'")
	if err != nil {
		out = "unavailable"
	}

	m := &models.ServerMetrics{
		ServerID:  a.serverID,
		Timestamp: time.Now(),
	}
	parseNGINXStatus(out, m)
	return m, nil
}

// parseNGINXStatus parses nginx stub_status output.
// Active connections: 291
// server accepts handled requests
//
//	16630948 16630948 31070465
//
// Reading: 6 Writing: 179 Waiting: 106
func parseNGINXStatus(raw string, m *models.ServerMetrics) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Active connections:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				v, _ := strconv.Atoi(parts[2])
				m.ActiveConnections = v
			}
		}
	}
}

func (a *NGINXAdapter) GetConfig(ctx context.Context) (*models.ProxyConfig, error) {
	client, err := a.getClient(ctx)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	content, err := runSession(session, "cat /etc/nginx/nginx.conf 2>/dev/null || cat /usr/local/etc/nginx/nginx.conf 2>/dev/null")
	if err != nil {
		return nil, err
	}

	return &models.ProxyConfig{
		ServerID:     a.serverID,
		ServerName:   a.serverName,
		ProxyType:    models.ProxyNGINX,
		Content:      content,
		Format:       "nginx",
		LastModified: time.Now().Format(time.RFC3339),
		IsValid:      true,
	}, nil
}

func (a *NGINXAdapter) PutConfig(ctx context.Context, content string) (*models.ConfigValidation, error) {
	client, err := a.getClient(ctx)
	if err != nil {
		return nil, err
	}

	escaped := strings.ReplaceAll(content, "'", "'\\''")
	writeCmd := fmt.Sprintf("echo '%s' > /tmp/nginx_proxera.conf && nginx -t -c /tmp/nginx_proxera.conf 2>&1", escaped)

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	out, err := runSession(session, writeCmd)
	if err != nil || strings.Contains(out, "failed") || strings.Contains(out, "[emerg]") {
		return &models.ConfigValidation{
			IsValid: false,
			Errors:  []string{out},
		}, nil
	}

	session2, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session2.Close()
	runSession(session2, "sudo cp /tmp/nginx_proxera.conf /etc/nginx/nginx.conf") //nolint:errcheck

	return &models.ConfigValidation{IsValid: true}, nil
}

func (a *NGINXAdapter) Reload(ctx context.Context) error {
	client, err := a.getClient(ctx)
	if err != nil {
		return err
	}
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	_, err = runSession(session, "sudo nginx -s reload")
	return err
}

func (a *NGINXAdapter) TailLogs(ctx context.Context) (io.ReadCloser, error) {
	client, err := a.getClient(ctx)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	session.Stdout = pw
	session.Stderr = pw

	if err := session.Start("tail -F /var/log/nginx/access.log /var/log/nginx/error.log 2>/dev/null"); err != nil {
		session.Close()
		pw.Close()
		return nil, err
	}

	go func() {
		<-ctx.Done()
		session.Close()
		pw.Close()
	}()

	go func() {
		session.Wait() //nolint:errcheck
		pw.Close()
	}()

	return pr, nil
}

func (a *NGINXAdapter) GetStatus(ctx context.Context) (string, error) {
	_, err := a.Ping(ctx)
	if err != nil {
		return "offline", nil
	}
	return "online", nil
}

// runSession runs a command in an SSH session and returns combined stdout+stderr output.
func runSession(session *ssh.Session, cmd string) (string, error) {
	var buf bytes.Buffer
	session.Stdout = &buf
	session.Stderr = &buf
	if err := session.Run(cmd); err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}
