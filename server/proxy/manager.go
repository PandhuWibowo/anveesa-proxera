package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/anveesa/proxera/models"
	"golang.org/x/crypto/ssh"
)

// SSHPool manages a pool of reusable SSH client connections.
type SSHPool struct {
	mu      sync.RWMutex
	clients map[string]*poolEntry
}

type poolEntry struct {
	client   *ssh.Client
	lastUsed time.Time
}

func NewSSHPool() *SSHPool {
	p := &SSHPool{
		clients: make(map[string]*poolEntry),
	}
	go p.evictLoop()
	return p
}

// Get returns an existing or new SSH client for the given server.
func (p *SSHPool) Get(ctx context.Context, serverID, host string, port int, user, privKeyPEM string) (*ssh.Client, error) {
	p.mu.RLock()
	entry, exists := p.clients[serverID]
	p.mu.RUnlock()

	if exists {
		// Validate connection with a test session
		sess, err := entry.client.NewSession()
		if err == nil {
			sess.Close()
			entry.lastUsed = time.Now()
			return entry.client, nil
		}
		// Connection stale — remove it
		p.Evict(serverID)
	}

	// Create new connection
	client, err := dialSSH(ctx, host, port, user, privKeyPEM)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.clients[serverID] = &poolEntry{client: client, lastUsed: time.Now()}
	p.mu.Unlock()

	go p.keepalive(serverID, client)

	return client, nil
}

// Evict removes and closes the pool entry for a given server.
func (p *SSHPool) Evict(serverID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.clients[serverID]; ok {
		entry.client.Close()
		delete(p.clients, serverID)
	}
}

func (p *SSHPool) evictLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		p.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for id, entry := range p.clients {
			if entry.lastUsed.Before(cutoff) {
				entry.client.Close()
				delete(p.clients, id)
				log.Printf("SSH pool: evicted idle connection for server %s", id)
			}
		}
		p.mu.Unlock()
	}
}

func (p *SSHPool) keepalive(serverID string, client *ssh.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		p.mu.RLock()
		entry, ok := p.clients[serverID]
		p.mu.RUnlock()
		if !ok || entry.client != client {
			return
		}
		if _, _, err := client.SendRequest("keepalive@proxera", true, nil); err != nil {
			p.Evict(serverID)
			return
		}
	}
}

func dialSSH(ctx context.Context, host string, port int, user, privKeyPEM string) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	if privKeyPEM != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // TODO: known_hosts in production
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	connCh := make(chan net.Conn, 1)
	errCh := make(chan error, 1)
	go func() {
		d := &net.Dialer{Timeout: 10 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			errCh <- err
			return
		}
		connCh <- conn
	}()

	var netConn net.Conn
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case netConn = <-connCh:
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(netConn, addr, cfg)
	if err != nil {
		netConn.Close()
		return nil, err
	}
	return ssh.NewClient(sshConn, chans, reqs), nil
}

// Manager builds the correct adapter for a server.
type Manager struct {
	sshPool *SSHPool
}

func NewManager() *Manager {
	return &Manager{sshPool: NewSSHPool()}
}

func (m *Manager) GetSSHPool() *SSHPool {
	return m.sshPool
}

// NewAdapter creates the appropriate ProxyAdapter for the given server config.
func (m *Manager) NewAdapter(
	serverID, serverName, host string,
	port int,
	proxyType, connectionType string,
	sshUser, sshKey, apiURL, apiToken string,
) (ProxyAdapter, error) {
	switch proxyType {
	case "nginx":
		return NewNGINXAdapter(serverID, serverName, host, port, sshUser, sshKey, m.sshPool), nil
	case "traefik":
		if apiURL == "" {
			apiURL = fmt.Sprintf("http://%s:%d", host, port)
		}
		return NewTraefikAdapter(serverID, serverName, apiURL, apiToken), nil
	case "caddy":
		if apiURL == "" {
			apiURL = fmt.Sprintf("http://%s:%d", host, port)
		}
		return NewCaddyAdapter(serverID, serverName, apiURL), nil
	case "haproxy":
		if apiURL == "" {
			apiURL = fmt.Sprintf("http://%s:%d", host, port)
		}
		return NewHAProxyAdapter(serverID, serverName, apiURL, apiToken), nil
	default:
		return &stubAdapter{serverID: serverID, host: host, port: port}, nil
	}
}

// stubAdapter is used for "other" proxy types.
type stubAdapter struct {
	serverID string
	host     string
	port     int
}

func (s *stubAdapter) Type() string { return "other" }

func (s *stubAdapter) Ping(ctx context.Context) (int64, error) {
	start := time.Now()
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start).Milliseconds(), nil
}

func (s *stubAdapter) GetMetrics(_ context.Context) (*models.ServerMetrics, error) {
	return &models.ServerMetrics{ServerID: s.serverID, Timestamp: time.Now()}, nil
}

func (s *stubAdapter) GetConfig(_ context.Context) (*models.ProxyConfig, error) {
	return nil, &ErrNotSupported{Op: "GetConfig"}
}

func (s *stubAdapter) PutConfig(_ context.Context, _ string) (*models.ConfigValidation, error) {
	return nil, &ErrNotSupported{Op: "PutConfig"}
}

func (s *stubAdapter) Reload(_ context.Context) error { return &ErrNotSupported{Op: "Reload"} }

func (s *stubAdapter) TailLogs(_ context.Context) (io.ReadCloser, error) {
	return nil, &ErrNotSupported{Op: "TailLogs"}
}

func (s *stubAdapter) GetStatus(ctx context.Context) (string, error) {
	_, err := s.Ping(ctx)
	if err != nil {
		return "offline", nil
	}
	return "online", nil
}
