package models

import (
	"time"
)

type ProxyType string
type ConnectionType string
type ServerStatus string

const (
	ProxyNGINX   ProxyType = "nginx"
	ProxyTraefik ProxyType = "traefik"
	ProxyCaddy   ProxyType = "caddy"
	ProxyHAProxy ProxyType = "haproxy"
	ProxyOther   ProxyType = "other"

	ConnSSH ConnectionType = "ssh"
	ConnAPI ConnectionType = "api"

	StatusOnline  ServerStatus = "online"
	StatusOffline ServerStatus = "offline"
	StatusWarning ServerStatus = "warning"
	StatusUnknown ServerStatus = "unknown"
)

type Server struct {
	ID             string         `gorm:"primaryKey;type:text" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Host           string         `gorm:"not null" json:"host"`
	Port           int            `json:"port"`
	ProxyType      ProxyType      `gorm:"not null" json:"proxyType"`
	ConnectionType ConnectionType `gorm:"not null" json:"connectionType"`
	Status         ServerStatus   `gorm:"default:unknown" json:"status"`
	Uptime         string         `json:"uptime"`
	Version        string         `json:"version"`
	Location       string         `json:"location,omitempty"`
	Description    string         `json:"description,omitempty"`
	TagsJSON       string         `gorm:"column:tags;default:'[]'" json:"-"`
	Tags           []string       `gorm:"-" json:"tags"`

	// SSH fields
	SSHUser       string `json:"sshUser,omitempty"`
	SSHKeyContent string `gorm:"column:ssh_key_enc" json:"-"` // stored encrypted

	// API fields
	APIURL       string `json:"apiUrl,omitempty"`
	APITokenEnc  string `gorm:"column:api_token_enc" json:"-"`      // stored encrypted
	APITokenMask string `gorm:"-" json:"apiToken,omitempty"`        // masked for read

	// Live metrics (not persisted)
	ActiveConnections int     `gorm:"-" json:"activeConnections"`
	RequestsPerSec    float64 `gorm:"-" json:"requestsPerSec"`
	ErrorRate         float64 `gorm:"-" json:"errorRate"`
	CPUUsage          float64 `gorm:"-" json:"cpuUsage"`
	MemUsage          float64 `gorm:"-" json:"memUsage"`

	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"-"`
	LastChecked *time.Time `json:"lastChecked,omitempty"`
}

type ServerMetrics struct {
	ServerID          string    `json:"serverId"`
	Timestamp         time.Time `json:"timestamp"`
	RequestsPerSec    float64   `json:"requestsPerSec"`
	ActiveConnections int       `json:"activeConnections"`
	ErrorRate         float64   `json:"errorRate"`
	P50Latency        float64   `json:"p50Latency"`
	P95Latency        float64   `json:"p95Latency"`
	P99Latency        float64   `json:"p99Latency"`
	CPUUsage          float64   `json:"cpuUsage"`
	MemUsage          float64   `json:"memUsage"`
	NetworkIn         float64   `json:"networkIn"`
	NetworkOut        float64   `json:"networkOut"`
}

type ProxyConfig struct {
	ServerID         string    `json:"serverId"`
	ServerName       string    `json:"serverName"`
	ProxyType        ProxyType `json:"proxyType"`
	Content          string    `json:"content"`
	Format           string    `json:"format"`
	LastModified     string    `json:"lastModified"`
	IsValid          bool      `json:"isValid"`
	ValidationErrors []string  `json:"validationErrors"`
}

type ConfigValidation struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"errors"`
}

type CreateServerRequest struct {
	Name           string         `json:"name" binding:"required"`
	Host           string         `json:"host" binding:"required"`
	Port           int            `json:"port"`
	ProxyType      ProxyType      `json:"proxyType" binding:"required"`
	ConnectionType ConnectionType `json:"connectionType" binding:"required"`
	Location       string         `json:"location"`
	Description    string         `json:"description"`
	Tags           []string       `json:"tags"`
	SSHUser        string         `json:"sshUser"`
	SSHKey         string         `json:"sshKey"`
	APIURL         string         `json:"apiUrl"`
	APIToken       string         `json:"apiToken"`
}

type UpdateServerRequest struct {
	Name           *string         `json:"name"`
	Host           *string         `json:"host"`
	Port           *int            `json:"port"`
	ProxyType      *ProxyType      `json:"proxyType"`
	ConnectionType *ConnectionType `json:"connectionType"`
	Location       *string         `json:"location"`
	Description    *string         `json:"description"`
	Tags           []string        `json:"tags"`
	SSHUser        *string         `json:"sshUser"`
	SSHKey         *string         `json:"sshKey"`
	APIURL         *string         `json:"apiUrl"`
	APIToken       *string         `json:"apiToken"`
}
