package models

import "time"

type LoadBalancingMethod string

const (
	LBRoundRobin LoadBalancingMethod = "round_robin"
	LBLeastConn  LoadBalancingMethod = "least_conn"
	LBIPHash     LoadBalancingMethod = "ip_hash"
	LBRandom     LoadBalancingMethod = "random"
)

type Route struct {
	ID                  string              `gorm:"primaryKey;type:text" json:"id"`
	ServerID            string              `gorm:"not null;index" json:"serverId"`
	Server              *Server             `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Name                string              `gorm:"not null" json:"name"`
	Enabled             bool                `gorm:"default:true" json:"enabled"`
	MatchHost           string              `json:"matchHost,omitempty"`
	MatchPath           string              `json:"matchPath,omitempty"`
	MatchMethod         string              `json:"matchMethod,omitempty"`
	TargetUpstream      string              `gorm:"not null" json:"targetUpstream"`
	LoadBalancingMethod LoadBalancingMethod `gorm:"default:round_robin" json:"loadBalancingMethod"`
	SSLEnabled          bool                `json:"sslEnabled"`
	SSLCertExpiry       *time.Time          `json:"sslCertExpiry,omitempty"`
	MiddlewaresJSON     string              `gorm:"column:middlewares;default:'[]'" json:"-"`
	Middlewares         []string            `gorm:"-" json:"middlewares"`
	Priority            int                 `gorm:"default:0" json:"priority"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
	DeletedAt           *time.Time          `gorm:"index" json:"-"`
}

type RouteResponse struct {
	ID                  string              `json:"id"`
	ServerID            string              `json:"serverId"`
	ServerName          string              `json:"serverName"`
	Name                string              `json:"name"`
	Enabled             bool                `json:"enabled"`
	MatchHost           string              `json:"matchHost,omitempty"`
	MatchPath           string              `json:"matchPath,omitempty"`
	MatchMethod         string              `json:"matchMethod,omitempty"`
	TargetUpstream      string              `json:"targetUpstream"`
	LoadBalancingMethod LoadBalancingMethod `json:"loadBalancingMethod"`
	SSLEnabled          bool                `json:"sslEnabled"`
	SSLCertExpiry       *time.Time          `json:"sslCertExpiry,omitempty"`
	Middlewares         []string            `json:"middlewares"`
	Priority            int                 `json:"priority"`
	CreatedAt           time.Time           `json:"createdAt"`
}

type CreateRouteRequest struct {
	ServerID            string              `json:"serverId" binding:"required"`
	Name                string              `json:"name" binding:"required"`
	Enabled             bool                `json:"enabled"`
	MatchHost           string              `json:"matchHost"`
	MatchPath           string              `json:"matchPath"`
	MatchMethod         string              `json:"matchMethod"`
	TargetUpstream      string              `json:"targetUpstream" binding:"required"`
	LoadBalancingMethod LoadBalancingMethod `json:"loadBalancingMethod"`
	SSLEnabled          bool                `json:"sslEnabled"`
	SSLCertExpiry       *time.Time          `json:"sslCertExpiry"`
	Middlewares         []string            `json:"middlewares"`
	Priority            int                 `json:"priority"`
}

type UpdateRouteRequest struct {
	Name                *string              `json:"name"`
	Enabled             *bool                `json:"enabled"`
	MatchHost           *string              `json:"matchHost"`
	MatchPath           *string              `json:"matchPath"`
	MatchMethod         *string              `json:"matchMethod"`
	TargetUpstream      *string              `json:"targetUpstream"`
	LoadBalancingMethod *LoadBalancingMethod `json:"loadBalancingMethod"`
	SSLEnabled          *bool                `json:"sslEnabled"`
	SSLCertExpiry       *time.Time           `json:"sslCertExpiry"`
	Middlewares         []string             `json:"middlewares"`
	Priority            *int                 `json:"priority"`
}
