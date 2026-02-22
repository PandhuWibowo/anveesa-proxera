package handlers

import (
	"net/http"
	"time"

	"github.com/anveesa/proxera/database"
	"github.com/anveesa/proxera/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListAlerts GET /api/v1/alerts
func ListAlerts(c *gin.Context) {
	var alerts []models.Alert
	q := database.DB.Order("created_at DESC")

	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if sev := c.Query("severity"); sev != "" {
		q = q.Where("severity = ?", sev)
	}
	if sid := c.Query("serverId"); sid != "" {
		q = q.Where("server_id = ?", sid)
	}

	if err := q.Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

// CreateAlert POST /api/v1/alerts
func CreateAlert(c *gin.Context) {
	var req models.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert := models.Alert{
		ID:         uuid.New().String(),
		ServerID:   req.ServerID,
		ServerName: req.ServerName,
		Severity:   req.Severity,
		Status:     models.AlertStatusActive,
		Title:      req.Title,
		Message:    req.Message,
		Category:   req.Category,
	}

	if err := database.DB.Create(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	Hub.BroadcastAlert(alert)

	c.JSON(http.StatusCreated, alert)
}

// GetAlert GET /api/v1/alerts/:id
func GetAlert(c *gin.Context) {
	alert, ok := findAlert(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, alert)
}

// UpdateAlert PATCH /api/v1/alerts/:id
func UpdateAlert(c *gin.Context) {
	alert, ok := findAlert(c)
	if !ok {
		return
	}

	var req models.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != nil {
		alert.Status = *req.Status
	}
	if req.Severity != nil {
		alert.Severity = *req.Severity
	}
	if req.Title != nil {
		alert.Title = *req.Title
	}
	if req.Message != nil {
		alert.Message = *req.Message
	}

	if err := database.DB.Save(alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alert)
}

// DeleteAlert DELETE /api/v1/alerts/:id
func DeleteAlert(c *gin.Context) {
	alert, ok := findAlert(c)
	if !ok {
		return
	}
	if err := database.DB.Delete(alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert deleted"})
}

// AcknowledgeAlert POST /api/v1/alerts/:id/acknowledge
func AcknowledgeAlert(c *gin.Context) {
	alert, ok := findAlert(c)
	if !ok {
		return
	}
	alert.Status = models.AlertStatusAcknowledged
	if err := database.DB.Save(alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alert)
}

// ResolveAlert POST /api/v1/alerts/:id/resolve
func ResolveAlert(c *gin.Context) {
	alert, ok := findAlert(c)
	if !ok {
		return
	}
	now := time.Now()
	alert.Status = models.AlertStatusResolved
	alert.ResolvedAt = &now
	if err := database.DB.Save(alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alert)
}

// BulkAcknowledgeAlerts POST /api/v1/alerts/bulk/acknowledge
func BulkAcknowledgeAlerts(c *gin.Context) {
	var req models.BulkAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&models.Alert{}).
		Where("id IN ?", req.IDs).
		Update("status", models.AlertStatusAcknowledged).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"acknowledged": len(req.IDs)})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func findAlert(c *gin.Context) (*models.Alert, bool) {
	id := c.Param("id")
	var alert models.Alert
	if err := database.DB.First(&alert, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return nil, false
	}
	return &alert, true
}
