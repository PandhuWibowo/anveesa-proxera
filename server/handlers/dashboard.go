package handlers

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/anveesa/proxera/database"
	"github.com/anveesa/proxera/models"
	"github.com/gin-gonic/gin"
)

// DashboardStats GET /api/v1/dashboard/stats
func DashboardStats(c *gin.Context) {
	var stats models.DashboardStats

	var total int64
	database.DB.Model(&models.Server{}).Where("deleted_at IS NULL").Count(&total)
	stats.TotalServers = int(total)

	var onlineCount, offlineCount int64
	database.DB.Model(&models.Server{}).Where("deleted_at IS NULL AND status = ?", "online").Count(&onlineCount)
	database.DB.Model(&models.Server{}).Where("deleted_at IS NULL AND status = ?", "offline").Count(&offlineCount)
	stats.OnlineServers = int(onlineCount)
	stats.OfflineServers = int(offlineCount)

	var routeCount int64
	database.DB.Model(&models.Route{}).Where("deleted_at IS NULL").Count(&routeCount)
	stats.TotalRoutes = int(routeCount)

	var alertCount int64
	database.DB.Model(&models.Alert{}).Where("status = ?", "active").Count(&alertCount)
	stats.ActiveAlerts = int(alertCount)

	// Synthetic aggregates — replace with real metrics store in production
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	stats.TotalRequestsToday = int64(rng.Intn(1000000) + 100000)
	stats.AvgErrorRate = rng.Float64() * 2
	stats.AvgLatency = rng.Float64()*50 + 10

	c.JSON(http.StatusOK, stats)
}

// DashboardTraffic GET /api/v1/dashboard/traffic
func DashboardTraffic(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 || hours > 168 {
		hours = 24
	}

	now := time.Now().Truncate(time.Hour)
	points := make([]models.TrafficPoint, hours)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	baseRequests := 5000
	for i := hours - 1; i >= 0; i-- {
		t := now.Add(time.Duration(-i) * time.Hour)
		hour := t.Hour()

		var multiplier float64
		switch {
		case hour >= 9 && hour <= 17:
			multiplier = 1.0 + rng.Float64()*0.5
		case hour >= 18 && hour <= 22:
			multiplier = 0.7 + rng.Float64()*0.3
		default:
			multiplier = 0.2 + rng.Float64()*0.2
		}

		requests := int(float64(baseRequests) * multiplier)
		errorRate := 0.01 + rng.Float64()*0.02
		points[hours-1-i] = models.TrafficPoint{
			Time:     t,
			Requests: requests,
			Errors:   int(float64(requests) * errorRate),
			Latency:  20 + rng.Float64()*30,
		}
	}

	c.JSON(http.StatusOK, points)
}
