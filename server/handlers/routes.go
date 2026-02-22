package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/anveesa/proxera/database"
	"github.com/anveesa/proxera/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListRoutes GET /api/v1/routes
func ListRoutes(c *gin.Context) {
	var routes []models.Route
	q := database.DB.Where("deleted_at IS NULL")

	if sid := c.Query("serverId"); sid != "" {
		q = q.Where("server_id = ?", sid)
	}

	if err := q.Find(&routes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]models.RouteResponse, 0, len(routes))
	for _, r := range routes {
		resp = append(resp, toRouteResponse(r))
	}
	c.JSON(http.StatusOK, resp)
}

// CreateRoute POST /api/v1/routes
func CreateRoute(c *gin.Context) {
	var req models.CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.LoadBalancingMethod == "" {
		req.LoadBalancingMethod = models.LBRoundRobin
	}

	middlewaresJSON := "[]"
	if req.Middlewares != nil {
		b, _ := json.Marshal(req.Middlewares)
		middlewaresJSON = string(b)
	}

	route := models.Route{
		ID:                  uuid.New().String(),
		ServerID:            req.ServerID,
		Name:                req.Name,
		Enabled:             req.Enabled,
		MatchHost:           req.MatchHost,
		MatchPath:           req.MatchPath,
		MatchMethod:         req.MatchMethod,
		TargetUpstream:      req.TargetUpstream,
		LoadBalancingMethod: req.LoadBalancingMethod,
		SSLEnabled:          req.SSLEnabled,
		SSLCertExpiry:       req.SSLCertExpiry,
		MiddlewaresJSON:     middlewaresJSON,
		Priority:            req.Priority,
	}

	if err := database.DB.Create(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toRouteResponse(route))
}

// GetRoute GET /api/v1/routes/:id
func GetRoute(c *gin.Context) {
	route, ok := findRoute(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, toRouteResponse(*route))
}

// UpdateRoute PUT /api/v1/routes/:id
func UpdateRoute(c *gin.Context) {
	route, ok := findRoute(c)
	if !ok {
		return
	}

	var req models.CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	route.Name = req.Name
	route.Enabled = req.Enabled
	route.MatchHost = req.MatchHost
	route.MatchPath = req.MatchPath
	route.MatchMethod = req.MatchMethod
	route.TargetUpstream = req.TargetUpstream
	route.LoadBalancingMethod = req.LoadBalancingMethod
	route.SSLEnabled = req.SSLEnabled
	route.SSLCertExpiry = req.SSLCertExpiry
	route.Priority = req.Priority

	if req.Middlewares != nil {
		b, _ := json.Marshal(req.Middlewares)
		route.MiddlewaresJSON = string(b)
	}

	if err := database.DB.Save(route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toRouteResponse(*route))
}

// PatchRoute PATCH /api/v1/routes/:id
func PatchRoute(c *gin.Context) {
	route, ok := findRoute(c)
	if !ok {
		return
	}

	var req models.UpdateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		route.Name = *req.Name
	}
	if req.Enabled != nil {
		route.Enabled = *req.Enabled
	}
	if req.MatchHost != nil {
		route.MatchHost = *req.MatchHost
	}
	if req.MatchPath != nil {
		route.MatchPath = *req.MatchPath
	}
	if req.MatchMethod != nil {
		route.MatchMethod = *req.MatchMethod
	}
	if req.TargetUpstream != nil {
		route.TargetUpstream = *req.TargetUpstream
	}
	if req.LoadBalancingMethod != nil {
		route.LoadBalancingMethod = *req.LoadBalancingMethod
	}
	if req.SSLEnabled != nil {
		route.SSLEnabled = *req.SSLEnabled
	}
	if req.SSLCertExpiry != nil {
		route.SSLCertExpiry = req.SSLCertExpiry
	}
	if req.Priority != nil {
		route.Priority = *req.Priority
	}
	if req.Middlewares != nil {
		b, _ := json.Marshal(req.Middlewares)
		route.MiddlewaresJSON = string(b)
	}

	if err := database.DB.Save(route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toRouteResponse(*route))
}

// DeleteRoute DELETE /api/v1/routes/:id
func DeleteRoute(c *gin.Context) {
	route, ok := findRoute(c)
	if !ok {
		return
	}

	now := time.Now()
	route.DeletedAt = &now
	if err := database.DB.Save(route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "route deleted"})
}

// ToggleRoute POST /api/v1/routes/:id/toggle
func ToggleRoute(c *gin.Context) {
	route, ok := findRoute(c)
	if !ok {
		return
	}

	route.Enabled = !route.Enabled
	if err := database.DB.Save(route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": route.ID, "enabled": route.Enabled})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func findRoute(c *gin.Context) (*models.Route, bool) {
	id := c.Param("id")
	var route models.Route
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(&route).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return nil, false
	}
	return &route, true
}

func toRouteResponse(r models.Route) models.RouteResponse {
	var middlewares []string
	if r.MiddlewaresJSON != "" {
		json.Unmarshal([]byte(r.MiddlewaresJSON), &middlewares) //nolint:errcheck
	}
	if middlewares == nil {
		middlewares = []string{}
	}

	serverName := ""
	if r.Server != nil {
		serverName = r.Server.Name
	}

	return models.RouteResponse{
		ID:                  r.ID,
		ServerID:            r.ServerID,
		ServerName:          serverName,
		Name:                r.Name,
		Enabled:             r.Enabled,
		MatchHost:           r.MatchHost,
		MatchPath:           r.MatchPath,
		MatchMethod:         r.MatchMethod,
		TargetUpstream:      r.TargetUpstream,
		LoadBalancingMethod: r.LoadBalancingMethod,
		SSLEnabled:          r.SSLEnabled,
		SSLCertExpiry:       r.SSLCertExpiry,
		Middlewares:         middlewares,
		Priority:            r.Priority,
		CreatedAt:           r.CreatedAt,
	}
}
