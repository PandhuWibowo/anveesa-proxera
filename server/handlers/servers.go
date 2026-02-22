package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anveesa/proxera/crypto"
	"github.com/anveesa/proxera/database"
	"github.com/anveesa/proxera/models"
	"github.com/anveesa/proxera/proxy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var proxyManager = proxy.NewManager()

// ListServers GET /api/v1/servers
func ListServers(c *gin.Context) {
	var servers []models.Server
	q := database.DB.Where("deleted_at IS NULL")

	if t := c.Query("type"); t != "" {
		q = q.Where("proxy_type = ?", t)
	}
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}

	if err := q.Find(&servers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range servers {
		unmarshalTags(&servers[i])
		if servers[i].APITokenEnc != "" {
			if dec, err := crypto.Decrypt(servers[i].APITokenEnc); err == nil {
				servers[i].APITokenMask = crypto.MaskToken(dec)
			}
		}
	}
	c.JSON(http.StatusOK, servers)
}

// GetServer GET /api/v1/servers/:id
func GetServer(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}
	unmarshalTags(server)
	if server.APITokenEnc != "" {
		if dec, err := crypto.Decrypt(server.APITokenEnc); err == nil {
			server.APITokenMask = crypto.MaskToken(dec)
		}
	}
	c.JSON(http.StatusOK, server)
}

// CreateServer POST /api/v1/servers
func CreateServer(c *gin.Context) {
	var req models.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Port == 0 {
		req.Port = defaultPort(string(req.ProxyType))
	}

	server := models.Server{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Host:           req.Host,
		Port:           req.Port,
		ProxyType:      req.ProxyType,
		ConnectionType: req.ConnectionType,
		Status:         models.StatusUnknown,
		Location:       req.Location,
		Description:    req.Description,
		SSHUser:        req.SSHUser,
		APIURL:         req.APIURL,
	}

	if req.Tags != nil {
		b, _ := json.Marshal(req.Tags)
		server.TagsJSON = string(b)
	} else {
		server.TagsJSON = "[]"
	}

	if req.SSHKey != "" {
		enc, err := crypto.Encrypt(req.SSHKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
			return
		}
		server.SSHKeyContent = enc
	}

	if req.APIToken != "" {
		enc, err := crypto.Encrypt(req.APIToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
			return
		}
		server.APITokenEnc = enc
	}

	if err := database.DB.Create(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	unmarshalTags(&server)
	c.JSON(http.StatusCreated, server)
}

// UpdateServer PUT /api/v1/servers/:id
func UpdateServer(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	var req models.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server.Name = req.Name
	server.Host = req.Host
	if req.Port != 0 {
		server.Port = req.Port
	}
	server.ProxyType = req.ProxyType
	server.ConnectionType = req.ConnectionType
	server.Location = req.Location
	server.Description = req.Description
	server.SSHUser = req.SSHUser
	server.APIURL = req.APIURL

	if req.Tags != nil {
		b, _ := json.Marshal(req.Tags)
		server.TagsJSON = string(b)
	}
	if req.SSHKey != "" {
		enc, _ := crypto.Encrypt(req.SSHKey)
		server.SSHKeyContent = enc
	}
	if req.APIToken != "" {
		enc, _ := crypto.Encrypt(req.APIToken)
		server.APITokenEnc = enc
	}

	proxyManager.GetSSHPool().Evict(server.ID)

	if err := database.DB.Save(server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	unmarshalTags(server)
	c.JSON(http.StatusOK, server)
}

// PatchServer PATCH /api/v1/servers/:id
func PatchServer(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	var req models.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		server.Name = *req.Name
	}
	if req.Host != nil {
		server.Host = *req.Host
	}
	if req.Port != nil {
		server.Port = *req.Port
	}
	if req.ProxyType != nil {
		server.ProxyType = *req.ProxyType
	}
	if req.ConnectionType != nil {
		server.ConnectionType = *req.ConnectionType
	}
	if req.Location != nil {
		server.Location = *req.Location
	}
	if req.Description != nil {
		server.Description = *req.Description
	}
	if req.SSHUser != nil {
		server.SSHUser = *req.SSHUser
	}
	if req.APIURL != nil {
		server.APIURL = *req.APIURL
	}
	if req.Tags != nil {
		b, _ := json.Marshal(req.Tags)
		server.TagsJSON = string(b)
	}
	if req.SSHKey != nil && *req.SSHKey != "" {
		enc, _ := crypto.Encrypt(*req.SSHKey)
		server.SSHKeyContent = enc
	}
	if req.APIToken != nil && *req.APIToken != "" {
		enc, _ := crypto.Encrypt(*req.APIToken)
		server.APITokenEnc = enc
	}

	proxyManager.GetSSHPool().Evict(server.ID)

	if err := database.DB.Save(server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	unmarshalTags(server)
	c.JSON(http.StatusOK, server)
}

// DeleteServer DELETE /api/v1/servers/:id
func DeleteServer(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	proxyManager.GetSSHPool().Evict(server.ID)

	now := time.Now()
	server.DeletedAt = &now
	if err := database.DB.Save(server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "server deleted"})
}

// ServerHealth GET /api/v1/servers/:id/health
func ServerHealth(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	adapter, err := buildAdapter(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	latency, pingErr := adapter.Ping(ctx)
	status := "online"
	if pingErr != nil {
		status = "offline"
	}

	now := time.Now()
	database.DB.Model(server).Updates(map[string]interface{}{
		"status":       status,
		"last_checked": now,
	})

	c.JSON(http.StatusOK, gin.H{
		"serverId":  server.ID,
		"status":    status,
		"latencyMs": latency,
		"checkedAt": now,
	})
}

// ServerMetrics GET /api/v1/servers/:id/metrics
func ServerMetrics(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	adapter, err := buildAdapter(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	metrics, err := adapter.GetMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetServerConfig GET /api/v1/servers/:id/config
func GetServerConfig(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	adapter, err := buildAdapter(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	cfg, err := adapter.GetConfig(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// PutServerConfig PUT /api/v1/servers/:id/config
func PutServerConfig(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	var body struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adapter, err := buildAdapter(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := adapter.PutConfig(ctx, body.Content)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ReloadServer POST /api/v1/servers/:id/reload
func ReloadServer(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	adapter, err := buildAdapter(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := adapter.Reload(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reload initiated"})
}

// StreamServerLogs GET /api/v1/servers/:id/logs (SSE)
func StreamServerLogs(c *gin.Context) {
	server, ok := findServer(c)
	if !ok {
		return
	}

	adapter, err := buildAdapter(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	rc, err := adapter.TailLogs(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer rc.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	fmt.Fprintf(c.Writer, "retry: 3000\n\n")
	c.Writer.Flush()

	scanner := bufio.NewScanner(rc)
	logID := 0
	for scanner.Scan() {
		line := scanner.Text()
		logID++
		entry := map[string]interface{}{
			"id":        fmt.Sprintf("l%d", logID),
			"serverId":  server.ID,
			"level":     parseLogLevel(line),
			"message":   line,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		data, _ := json.Marshal(entry)
		fmt.Fprintf(c.Writer, "id: l%d\nevent: log\ndata: %s\n\n", logID, string(data))
		c.Writer.Flush()

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func findServer(c *gin.Context) (*models.Server, bool) {
	id := c.Param("id")
	var server models.Server
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(&server).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return nil, false
	}
	return &server, true
}

func unmarshalTags(s *models.Server) {
	if s.TagsJSON != "" {
		json.Unmarshal([]byte(s.TagsJSON), &s.Tags) //nolint:errcheck
	}
	if s.Tags == nil {
		s.Tags = []string{}
	}
}

func buildAdapter(s *models.Server) (proxy.ProxyAdapter, error) {
	var sshKey, apiToken string
	if s.SSHKeyContent != "" {
		dec, err := crypto.Decrypt(s.SSHKeyContent)
		if err != nil {
			return nil, fmt.Errorf("decrypt ssh key: %w", err)
		}
		sshKey = dec
	}
	if s.APITokenEnc != "" {
		dec, err := crypto.Decrypt(s.APITokenEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt api token: %w", err)
		}
		apiToken = dec
	}
	return proxyManager.NewAdapter(
		s.ID, s.Name, s.Host, s.Port,
		string(s.ProxyType), string(s.ConnectionType),
		s.SSHUser, sshKey, s.APIURL, apiToken,
	)
}

func defaultPort(proxyType string) int {
	switch proxyType {
	case "nginx":
		return 22 // SSH port for NGINX
	case "traefik":
		return 8080
	case "caddy":
		return 2019
	case "haproxy":
		return 9090
	default:
		return 80
	}
}

func parseLogLevel(line string) string {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "crit") || strings.Contains(lower, "emerg"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warn"
	case strings.Contains(lower, "debug"):
		return "debug"
	default:
		return "info"
	}
}
