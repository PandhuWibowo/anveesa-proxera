package models

import "time"

type AlertSeverity string
type AlertStatus string
type AlertCategory string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityWarning  AlertSeverity = "warning"
	SeverityInfo     AlertSeverity = "info"

	AlertStatusActive       AlertStatus = "active"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusAcknowledged AlertStatus = "acknowledged"

	CategoryDowntime    AlertCategory = "downtime"
	CategoryConfig      AlertCategory = "config"
	CategorySSL         AlertCategory = "ssl"
	CategoryPerformance AlertCategory = "performance"
	CategorySecurity    AlertCategory = "security"
)

type Alert struct {
	ID         string        `gorm:"primaryKey;type:text" json:"id"`
	ServerID   string        `gorm:"index" json:"serverId,omitempty"`
	ServerName string        `json:"serverName,omitempty"`
	Severity   AlertSeverity `gorm:"not null" json:"severity"`
	Status     AlertStatus   `gorm:"default:active" json:"status"`
	Title      string        `gorm:"not null" json:"title"`
	Message    string        `json:"message"`
	Category   AlertCategory `json:"category"`
	CreatedAt  time.Time     `json:"timestamp"`
	UpdatedAt  time.Time     `json:"updatedAt"`
	ResolvedAt *time.Time    `json:"resolvedAt,omitempty"`
}

type CreateAlertRequest struct {
	ServerID   string        `json:"serverId"`
	ServerName string        `json:"serverName"`
	Severity   AlertSeverity `json:"severity" binding:"required"`
	Title      string        `json:"title" binding:"required"`
	Message    string        `json:"message"`
	Category   AlertCategory `json:"category"`
}

type UpdateAlertRequest struct {
	Status   *AlertStatus   `json:"status"`
	Severity *AlertSeverity `json:"severity"`
	Title    *string        `json:"title"`
	Message  *string        `json:"message"`
}

type BulkAlertRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type TrafficPoint struct {
	Time     time.Time `json:"time"`
	Requests int       `json:"requests"`
	Errors   int       `json:"errors"`
	Latency  float64   `json:"latency"`
}

type DashboardStats struct {
	TotalServers       int     `json:"totalServers"`
	OnlineServers      int     `json:"onlineServers"`
	OfflineServers     int     `json:"offlineServers"`
	TotalRoutes        int     `json:"totalRoutes"`
	ActiveAlerts       int     `json:"activeAlerts"`
	TotalRequestsToday int64   `json:"totalRequestsToday"`
	AvgErrorRate       float64 `json:"avgErrorRate"`
	AvgLatency         float64 `json:"avgLatency"`
}
